/*
   Copyright 2015 Daniel Gruber, Univa, My blog: www.gridengine.eu

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package staging

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgruber/ubercluster/pkg/http_helper"
	"github.com/dgruber/ubercluster/pkg/output"
	"github.com/dgruber/ubercluster/pkg/types"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type Filesystem struct {
	client *http.Client
}

func NewFilesystem(client *http.Client) *Filesystem {
	return &Filesystem{
		client: client,
	}
}

// CheckUploadFilesystem pre-checks the configured directory which
// is going to used for files staging during startup of the proxy
func CheckUploadFilesystem(dirname string) error {
	// check if directory exists
	if fi, err := os.Stat(dirname); err == nil {
		if fi.IsDir() {
			return nil
		}
		log.Println("Error: A file with same name than upload directory exists: ", dirname)
		return errors.New("File with same name as upload directory exists...")
	} else {
		if os.IsNotExist(err) {
			// create it
			log.Println("Creating file upload directory: ", dirname)
			return os.Mkdir(dirname, 0700)
		}
		return err
	}
}

// Client functionality

func fileUpload(url string, params map[string]string, paramName, filePath string) (*http.Request, error) {
	var err error
	var file *os.File

	if file, err = os.Open(filePath); err != nil {
		log.Println("Error when opening local file: ", nil)
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile(paramName, filepath.Base(filePath))
	if err != nil {
		return nil, err
	}

	if _, err = io.Copy(part, file); err != nil {
		log.Println("fileUpload copy error", err)
		return nil, err
	} else {
		for key, val := range params {
			_ = writer.WriteField(key, val)
		}
		if err = writer.Close(); err != nil {
			log.Println("fileUpload writer close error", err)
			return nil, err
		}
	}

	if req, reqErr := http.NewRequest("POST", url, body); reqErr == nil {
		req.Header.Add("Content-Type", writer.FormDataContentType())
		return req, nil
	} else {
		return req, reqErr
	}
}

// FsUploadFile uploads a file given by the path to a given
// cluster by setting a security key if required.
func (fs *Filesystem) FsUploadFile(otp, clusteraddress, jsName, filename string) {
	if filename == "" {
		fmt.Println("No filename given.")
		return // nothing to do
	}
	url := fmt.Sprintf("%s/jsession/%s/staging/upload", clusteraddress, jsName)
	log.Println("Created url: ", url)
	params := make(map[string]string)
	params["permission"] = "exec"
	// set otp
	if otp != "" {
		params["otp"] = otp
	}

	if req, err := fileUpload(url, params, "file", filename); err != nil {
		fmt.Println("Error during filupload: ", err)
		os.Exit(2)
	} else {
		log.Println("Request: ", req)
		if r, err := fs.client.Do(req); err == nil {
			r.Body.Close()
			fmt.Println("Uploaded file ", filename, r.Status)
		} else {
			fmt.Println("Error during file upload: ", err)
		}
	}
}

// UC fs interface

// fsListFiles requests a list of files from the given
// cluster and displays it
func getFiles(client *http.Client, otp, clusteraddress, jsName string) ([]types.FileInfo, error) {
	request := fmt.Sprintf("%s/jsession/%s/staging/files", clusteraddress, jsName)
	log.Println("Requesting:" + request)
	resp, err := http_helper.UberGet(client, otp, request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var fileinfo []types.FileInfo
	if err := decoder.Decode(&fileinfo); err != nil {
		return fileinfo, err
	}
	return fileinfo, nil
}

// FsListFiles lists all files on the remote staging area,
// theirs sizes, and if they are executable (i.e. can run
// as remote jobs).
func (fs *Filesystem) FsListFiles(otp, clusteraddress, jsName string, of output.OutputFormater) {
	if fi, err := getFiles(fs.client, otp, clusteraddress, jsName); err != nil {
		fmt.Println("Error during fetching files in staging area: ", err)
		os.Exit(1)
	} else {
		// output the files in the given interface
		of.PrintFiles(fi)
	}
}

// fsUploadFiles uploads a given list of files to the
// given cluster's staging area
func (fs *Filesystem) FsUploadFiles(otp, clusteraddress, jsName string, files []string, of output.OutputFormater) {
	log.Println("Uploading following files: ", files)
	for _, file := range files {
		fs.FsUploadFile(otp, clusteraddress, jsName, file)
	}
}

func (fs *Filesystem) DownloadFile(otp, clusteraddress, jsName, file string) {
	url := fmt.Sprintf("%s/jsession/%s/staging/file/%s", clusteraddress, jsName, file)
	log.Println("Using url: ", url)
	if f, err := os.Create(file); err != nil {
		fmt.Println("Error during creation of file: ", err)
		os.Exit(1)
	} else {
		defer f.Close()
		if response, err := fs.client.Get(url); err != nil {
			fmt.Println("Error during fetching file: ", err)
			os.Exit(1)
		} else {
			defer response.Body.Close()
			fmt.Println("Copy file now...")
			size, err := io.Copy(f, response.Body)
			if err != nil {
				fmt.Println("Error while downloading", url, "-", err)
				return
			}
			fmt.Printf("Downloaded file %s (%d bytes)\n", file, size)
		}
	}
}

// FsDownloadFiles downloads a list list of files from a
// the staging area of a given cluster
func (fs *Filesystem) FsDownloadFiles(otp, clusteraddress, jsName string, files []string, of output.OutputFormater) {
	log.Println("Downloading following files: ", files)
	for _, file := range files {
		fs.DownloadFile(otp, clusteraddress, jsName, file)
	}
}
