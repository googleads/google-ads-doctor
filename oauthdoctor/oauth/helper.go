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

// Package oauth implements functions to diagnose the supported OAuth2 flows
// (web and installed app flows) in a Google Ads API client library client
// environment.
package oauth

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/googleads/google-ads-doctor/oauthdoctor/diag"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// This is a list of error codes (not comprehensive) returned by Google OAuth2
// endpoint based on Google Ads API scope.
const (
	AccessNotPermittedForManagerAccount = iota
	GoogleAdsAPIDisabled
	InvalidClientInfo
	InvalidRefreshToken
	InvalidCustomerID
	MissingDevToken
	Unauthenticated
	Unauthorized
	UnknownError

	GoogleAdsApiScope = "https://www.googleapis.com/auth/adwords"
)

// Config is a required configuration for diagnosing the OAuth2 flow based on
// the client library configuration.
type Config struct {
	ConfigFile diag.ConfigFile
	CustomerID string
	OAuthType  string
	Verbose    bool
}

// ConfigWriter allows replacement of key by a given value in a configuration.
type ConfigWriter interface {
	ReplaceConfig(k, v string) string
}

var (
	appVersion string

	stdinSanitizer = strings.NewReplacer("\n", "")

	readStdin = func() string {
		reader := bufio.NewReader(os.Stdin)
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading input (%s) from command line: %s", str, err)
		}

		return strings.TrimSpace(stdinSanitizer.Replace(str))
	}
)

// SimulateOAuthFlow simulates the OAuth2 flows supported by the Google Ads API
// client libraries.
func (c *Config) SimulateOAuthFlow() {
	switch c.OAuthType {
	case diag.Web:
		c.simulateWebFlow()
	case diag.InstalledApp:
		c.simulateAppFlow()
	case diag.ServiceAccount:
		c.simulateServiceAccFlow()
	}
}

// decodeError checks the JSON response in the error and determines the error
// code.
func (c *Config) decodeError(err error) int32 {
	errstr := err.Error()

	if strings.Contains(errstr, "invalid_client") {
		// Client ID and/or secret is invalid
		return InvalidClientInfo
	}
	if strings.Contains(errstr, "unauthorized_client") {
		// The given refresh token may not be generated with the given client ID
		// and secret
		return Unauthorized
	}
	if strings.Contains(errstr, "invalid_grant") {
		// Refresh token is not valid for any users
		return InvalidRefreshToken
	}
	if strings.Contains(errstr, "refresh token is not set") {
		return InvalidRefreshToken
	}
	if strings.Contains(errstr, "USER_PERMISSION_DENIED") {
		// User doesn't have permission to access Google Ads account
		return InvalidRefreshToken
	}
	if strings.Contains(errstr, "\"PERMISSION_DENIED\"") {
		return GoogleAdsAPIDisabled
	}
	if strings.Contains(errstr, "UNAUTHENTICATED") {
		return Unauthenticated
	}
	if strings.Contains(errstr, "CANNOT_BE_EXECUTED_BY_MANAGER_ACCOUNT") {
		// Request cannot be executed by a manager account
		return AccessNotPermittedForManagerAccount
	}
	if strings.Contains(errstr, "DEVELOPER_TOKEN_PARAMETER_MISSING") {
		return MissingDevToken
	}
	if strings.Contains(errstr, "INVALID_CUSTOMER_ID") {
		return InvalidCustomerID
	}
	return UnknownError
}

// diagnose handles the error by guiding the user to take appropriate
// actions to fix the OAuth2 error based on the error code.
func (c *Config) diagnose(err error) {
	// Print the given message from JSON response if there's any
	var parsedMsg map[string]interface{}
	if err := json.Unmarshal([]byte(err.Error()), &parsedMsg); err == nil {
		errMsg := parsedMsg["error"].(map[string]interface{})["message"]
		log.Print("JSON response error: " + errMsg.(string))
	}

	switch c.decodeError(err) {
	case AccessNotPermittedForManagerAccount:
		log.Print("ERROR: Your credentials are not permitted to access to a manager account." +
			"\nPlease create your credentials with a Google Ads account with manager access.")
	case GoogleAdsAPIDisabled:
		log.Print("Press <Enter> to continue after you enable Google Ads API")
		readStdin()
	case InvalidClientInfo:
		log.Print("ERROR: Your client ID and/or client secret may be invalid.")
		replaceCloudCredentials(&c.ConfigFile)
	case InvalidRefreshToken, Unauthorized:
		log.Print("ERROR: Your refresh token may be invalid.")
	case MissingDevToken:
		log.Print("ERROR: Your developer token is missing in the configuration file")
		replaceDevToken(&c.ConfigFile)
	case Unauthenticated:
		log.Print("ERROR: The login email may not have access to the given account.")
	case InvalidCustomerID:
		log.Print("ERROR: You customer ID is invalid.")
	default:
		var helperText string
		switch c.ConfigFile.OAuthType {
		case diag.ServiceAccount:
			helperText = "Please verify the path of JSON key file and impersonate email (or delegated email)."
		case diag.Web:
			helperText = "Please verify your developer token, client ID and client secret."
		case diag.InstalledApp:
			helperText = "Please verify your developer token, client ID, client secret and refresh token."
		}
		log.Print("ERROR: Your credentials are invalid but we cannot determine the exact error. " + helperText)
	}
}

var (
	getClientID = func() string {
		fmt.Print("New Client ID >> ")
		return readStdin()
	}

	getClientSecret = func() string {
		fmt.Print("New Client Secret >> ")
		return readStdin()
	}
)

// replaceCloudCredentials prompts the user to create a new client ID and
// secret and to then enter them at the prompt. The values entered will
// replace the existing values in the client library configuration file.
func replaceCloudCredentials(c ConfigWriter) {
	log.Print("Follow this guide to setup your OAuth2 client ID and client secret: " +
		"https://developers.google.com/adwords/api/docs/guides/first-api-call#set_up_oauth2_authentication")

	clientID := getClientID()
	clientSecret := getClientSecret()

	c.ReplaceConfig(diag.ClientID, clientID)
	c.ReplaceConfig(diag.ClientSecret, clientSecret)
}

// replaceDevToken guides the user to retrieve their developer token and
// enter it at the prompt. The entered value will replace the existing
// developer token in the client library configuration file.
var replaceDevToken = func(c ConfigWriter) {
	log.Print("Please follow this guide to retrieve your developer token: " +
		"https://developers.google.com/adwords/api/docs/guides/signup#step-2")
	log.Print("Pleae enter a new Developer Token here and it will replace " +
		"the one in your client library configuration file")

	fmt.Print("New Developer Token >> ")
	devToken := readStdin()

	c.ReplaceConfig(diag.DevToken, devToken)
}

// replaceRefreshToken asks the user if they want to replace the refresh
// token in the configuration file with the newly generated value.
func replaceRefreshToken(c ConfigWriter, refreshToken string) {
	log.Print("Would you like to replace your refresh token in the " +
		"client library config file with the new one generated?")

	fmt.Print("Enter Y for Yes [Anything else is No] >> ")
	answer := readStdin()

	if answer == "Y" {
		c.ReplaceConfig(diag.RefreshToken, refreshToken)
	} else {
		log.Print("Refresh token is NOT replaced")
	}
}

var oauthEndpoint = google.Endpoint

// oauth2Conf creates a corresponding OAuth2 config struct based on the
// given configuration details. This is only applicable when a refresh token
// is not given.
func (c *Config) oauth2Conf(redirectURL string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     c.ConfigFile.ConfigKeys.ClientID,
		ClientSecret: c.ConfigFile.ClientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{GoogleAdsApiScope},
		Endpoint:     oauthEndpoint,
	}
}

// Given the auth code returned after the authentication and authorization
// step, oauth2Client creates a HTTP client with an authorized access token.
func (c *Config) oauth2Client(code string) (*http.Client, string) {
	conf := c.oauth2Conf(InstalledAppRedirectURL)
	// Handle the exchange code to initiate a transport.
	token, err := conf.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatal(err)
	}
	return conf.Client(oauth2.NoContext, token), token.RefreshToken
}

var apiURL = "https://googleads.googleapis.com/v5/customers/"

// getAccount makes a HTTP request to Google Ads API customer account
// endpoint and parses the JSON response.
func (c *Config) getAccount(client *http.Client) (*bytes.Buffer, error) {
	req, err := http.NewRequest("GET", apiURL+c.CustomerID, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("user-agent", userAgent())
	req.Header.Set("developer-token", c.ConfigFile.DevToken)
	if c.ConfigFile.LoginCustomerID != "" {
		req.Header.Set("login-customer-id", c.ConfigFile.LoginCustomerID)
	}

	if c.Verbose {
		dump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Printf("Error printing HTTP request: %s", err)
		}
		log.Printf("Making a HTTP Request to Google Ads API:\n%v\n", c.sanitizeOutput(string(dump)))
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)

	var jsonBody map[string]interface{}
	json.Unmarshal(buf.Bytes(), &jsonBody)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("A HTTP Status (%s) is returned while calling %s", resp.Status, apiURL+c.CustomerID)
	}

	if jsonBody["error"] != nil {
		return nil, fmt.Errorf(jsonBody["error"].(string))
	}

	return buf, nil
}

// userAgent returns a User-Agent HTTP header for this tool.
func userAgent() string {
	ua := "google-ads-doctor/"
	if appVersion != "" {
		ua += appVersion
	} else {
		ua += "source"
	}
	return ua
}

func (c *Config) sanitizeOutput(s string) string {
	return strings.ReplaceAll(s, c.ConfigFile.DevToken, "REDACTED")
}

// ReadCustomerID retrieves the CID from stdin.
func ReadCustomerID() string {
	for {
		log.Print("Please enter a Google Ads account ID:")
		customerID := readStdin()

		if customerID != "" {
			return strings.ReplaceAll(customerID, "-", "")
		}
	}
}
