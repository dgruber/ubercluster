/*
   Copyright 2015 Daniel Gruber, Univa, My blog: http://www.gridengine.eu

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

package http_helper

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

// uberGet makes an http GET request. Depending on the uc
// configuration (currently cli param) it adds a one time
// password.
func UberGet(otp, request string) (resp *http.Response, err error) {
	newRequest := request
	if otp != "" {
		// adding http secret key (OTP)
		if strings.Contains(request, "?") {
			newRequest = fmt.Sprintf("%s&otp=%s", request, otp)
		} else {
			newRequest = fmt.Sprintf("%s?otp=%s", request, otp)
		}
	}
	log.Println("New request: ", newRequest)
	return http.Get(newRequest)
}

// uberPost is a http.Post replacement which adds otp requests
// and possibly others depending on the configuration.
func UberPost(otp, url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	newUrl := url
	if otp != "" {
		// adding http secret key (OTP)
		if strings.Contains(url, "?") {
			newUrl = fmt.Sprintf("%s&otp=%s", url, otp)
		} else {
			newUrl = fmt.Sprintf("%s?otp=%s", url, otp)
		}
	}
	log.Println("New POST: ", newUrl)
	return http.Post(newUrl, bodyType, body)
}
