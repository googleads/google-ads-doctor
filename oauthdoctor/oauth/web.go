// Copyright 2019 Google LLC
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package oauth contains functions that are specific to web OAuth flow. The web
// flow initially prompts user to login, grant permission, and redirects
// user back to the redirect URL specified in Google Cloud project.
package oauth

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/oauth2"
)

var authCode = make(chan string)

// simulateWebFlow simulates the web flow to see if it succeeds
// or fails. If it fails, it will try to examine the error and prompt user
// to fix it. Then it retries to connect again and prints the result of the
// 2nd attempt.
func (c *Config) simulateWebFlow() {
	// Can only register the handle once
	http.HandleFunc("/", serverHandler)

	accountInfo, err := c.connectWebFlow()

	if err != nil {
		if c.Verbose {
			log.Print(err)
		}
		c.diagnose(err)
		accountInfo, err = c.connectWebFlow()
	}

	close(authCode)

	if err == nil {
		if c.Verbose {
			log.Print(accountInfo.String())
		}
		log.Println("SUCCESS: OAuth test passed with given config file settings.")
	} else {
		if c.Verbose {
			log.Println(err)
		}
		log.Println("ERROR: OAuth test failed.")
	}
}

// connectWebFlow connects with web flow OAuth2 and starts a web server in the
// background. The parent process interacts with users on the command line,
// while the background process is waiting for the auth code returned
// after the authentication and authorization step. Once the auth code is
// received in the background process, the command line will continue the
// simulation process.
func (c *Config) connectWebFlow() (*bytes.Buffer, error) {
	log.Print("You will need to enter the URL http://localhost:8080 as a valid " +
		"redirect URI in your Google APIs Console's project (https://console.developers.google.com/apis/library). " +
		"Please follow this guide (https://developers.google.com/google-ads/api/docs/oauth/cloud-project) " +
		"for further instructions.")
	conf := c.oauth2Conf("http://localhost:8080")

	// Redirect user to Google's consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	log.Printf("Visit the URL for the auth dialog:\n%s\n", url)

	srv := runServer()

	code := <-authCode

	srv.Shutdown(context.Background())

	client, _ := c.oauth2Client(code)
	return c.getAccount(client)
}

// runServer starts a HTTP server as a background process.
func runServer() *http.Server {
	log.Print("Running HTTP server in the background at port 8080...")
	srv := &http.Server{Addr: ":8080"}
	go srv.ListenAndServe()
	return srv
}

// serverHandler handles all the HTTP home page requests. It parses the auth
// code and sends it to the channel, so the parent process can continue the
// simulation at the command line.
func serverHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")

	if code != "" {
		authCode <- code
		log.Print("OAuth code received by the HTTP server handler: " + code)
		fmt.Fprintf(w, "Auth code received")
	}
}
