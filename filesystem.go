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

package ubercluster

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

// FileInfo describes a file in the staging area
type FileInfo struct {
	Filename   string `json:"filename"`
	Bytes      int64  `json:"bytes"`
	Executable bool   `json:"executable`
}

// checkUploadFilesystem pre-checks the configured directory which
// is going to used for files staging during startup of the proxy
func checkUploadFilesystem(dirname string) error {
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

// uploadFile uploads a file given by the path to a given
// cluster by setting a security key if required.
func UploadFile(otp, clusteraddress, filename string) {
	if filename == "" {
		return // nothing to do
	}
	url := fmt.Sprintf("%s/ubercluster/fileupload", clusteraddress)
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
		var client http.Client
		if r, err := client.Do(req); err == nil {
			r.Body.Close()
			fmt.Println("Uploaded file ", filename, r.Status)
		}
	}
}

// UC fs interface

// fsListFiles requests a list of files from the given
// cluster and displays it
func getFiles(otp, clusteraddress string) ([]FileInfo, error) {
	request := fmt.Sprintf("%s%s", clusteraddress, "/jsession/staging/files")
	log.Println("Requesting:" + request)
	resp, err := UberGet(otp, request)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	var fileinfo []FileInfo
	if err := decoder.Decode(&fileinfo); err != nil {
		return fileinfo, err
	}
	return fileinfo, nil
}

func FsListFiles(otp, clusteraddress string, of OutputFormater) {
	if fi, err := getFiles(otp, clusteraddress); err != nil {
		fmt.Println("Error during fetching files in staging area: ", err)
		os.Exit(1)
	} else {
		// output the files in the given interface
		of.PrintFiles(fi)
	}
}

// fsUploadFiles uploads a given list of files to the
// given cluster's staging area
func FsUploadFiles(otp, clusteraddress string, files []string, of OutputFormater) {

}

// fsDownloadFiles downloads a list list of files from a
// the staging area of a given cluster
func FsDownloadFiles(otp, clusteraddress string, files []string, of OutputFormater) {

}
