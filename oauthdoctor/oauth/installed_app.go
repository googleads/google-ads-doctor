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

// This file contains functions that are specific to installed app OAuth
// flow.

import (
	"bytes"
	"fmt"
	"log"
	"runtime"

	"golang.org/x/oauth2"
)

const (
	// InstalledAppRedirectURL is the redirect URL for the web flow.
	InstalledAppRedirectURL = "urn:ietf:wg:oauth:2.0:oob"
)

// This function simulates the installed app flow to see if it succeeds
// or fails. If it fails, it will try to examine the error and prompt user
// to fix it. Then it retries to connect again and prints the result of the
// 2nd attempt.
func (c *Config) simulateAppFlow() {
	var refreshToken string

	accountInfo, err := c.connectWithRefreshToken()
	if err != nil {
		if c.Verbose {
			log.Print(err)
		}
		c.diagnose(err)
		accountInfo, refreshToken, err = c.reconnect(err)
	}

	if err == nil {
		if c.Verbose {
			log.Print(accountInfo)
		}
		log.Println("SUCCESS: OAuth test passed with given config file settings.")

		if refreshToken != "" {
			replaceRefreshToken(&c.ConfigFile, refreshToken)
		}
	} else {
		if c.Verbose {
			log.Println(err)
		}
		log.Println("ERROR: OAuth test failed.")
	}
}

// This function connects with OAuth2 based on the given error and then
// sends a HTTP request to Google Ads API to get account info.
func (c *Config) reconnect(err error) (*bytes.Buffer, string, error) {
	switch c.decodeError(err) {
	case GoogleAdsAPIDisabled:
		accountInfo, oErr := c.connectWithRefreshToken()
		return accountInfo, "", oErr
	case InvalidCustomerID:
		c.CustomerID = ReadCustomerID()
		accountInfo, oErr := c.connectWithRefreshToken()
		return accountInfo, "", oErr
	case InvalidClientInfo:
		accountInfo, oErr := c.connectWithRefreshToken()
		return accountInfo, "", oErr
	case AccessNotPermittedForManagerAccount:
		log.Print("Attempting to regenerate refresh token...")
		return c.connectWithNoRefreshToken()
	case InvalidRefreshToken:
		log.Print("Attempting to regenerate refresh token...")
		return c.connectWithNoRefreshToken()
	case MissingDevToken:
		accountInfo, oErr := c.connectWithRefreshToken()
		return accountInfo, "", oErr
	default:
		log.Print("Attempting to regenerate refresh token...")
		return c.connectWithNoRefreshToken()
	}
}

// This function simulates the auth code generation step during the OAuth2
// authentication and authorization step.
func (c *Config) genAuthCode() string {
	conf := c.oauth2Conf(InstalledAppRedirectURL)

	// Redirect the user to Google's consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	log.Printf("Visit the URL for the auth dialog:\n%s\n", url)

	log.Print(genAuthCodePrompt(runtime.GOOS))
	fmt.Print("Enter Code >> ")

	return readStdin()
}

// genAuthCodePrompt returns the operating specific command prompt.
func genAuthCodePrompt(goos string) string {
	var msg string

	if goos == "windows" {
		msg += "You are running Windows, so to properly copy and paste the URL "
		msg += "into the command prompt:\n"
		msg += "1) Ensure that 'Quick Edit' mode is ON for your Command Prompt\n"
		msg += "2) Hold down the shift key\n"
		msg += "3) Highlight the URL\n"
		msg += "4) Right click on the highlighted area\n"
	}
	msg += "Copy the code here to continue:"
	return msg
}

// This function prompts the user to login, gets the auth code and exchange
// for the refresh token. And then, it gets the account info. This function
// is used based on the assumption of missing/incorrect refresh token in the
// client library config file.
func (c *Config) connectWithNoRefreshToken() (*bytes.Buffer, string, error) {
	code := c.genAuthCode()
	client, refreshToken := c.oauth2Client(code)
	accountInfo, err := c.getAccount(client)
	return accountInfo, refreshToken, err
}

// With refresh token given from client lib config file, it directly connects
// with OAuth and get the account info.
func (c *Config) connectWithRefreshToken() (*bytes.Buffer, error) {
	conf := &oauth2.Config{
		ClientID:     c.ConfigFile.ConfigKeys.ClientID,
		ClientSecret: c.ConfigFile.ClientSecret,
		Endpoint:     oauthEndpoint,
	}
	token := &oauth2.Token{RefreshToken: c.ConfigFile.RefreshToken}
	client := conf.Client(oauth2.NoContext, token)

	return c.getAccount(client)
}
