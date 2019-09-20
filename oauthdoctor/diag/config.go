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

// Package diag implements functions to diagnose a Google Ads client environment.
package diag

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/structs"
)

const (
	// DevToken is an abbreviation for developer token.
	// https://developers.google.com/google-ads/api/docs/first-call/dev-token
	DevToken = "DevToken"
	// ClientID is the OAuth client ID.
	ClientID = "ClientID"
	// ClientSecret is the secret provided from the Google API Console.
	// https://developers.google.com/google-ads/api/docs/oauth/cloud-project#create_a_client_id_and_client_secret/
	ClientSecret = "ClientSecret"
	// RefreshToken allows the client to obtain a new access token.
	RefreshToken = "RefreshToken"
)

var (
	// PIIWords is a slice of constant strings that indicate Personally Identifiable Information
	PIIWords = []string{DevToken, ClientID, ClientSecret, RefreshToken}

	// RequiredKeys are the key names used in the Language structure that defines
	// the contents of a client library configuration file.
	RequiredKeys = []string{DevToken, ClientID, ClientSecret, RefreshToken}
)

// Config is the collection of language specific elements.
type Config struct {
	Comment
	Separator string
	Cfg       ConfigFile
}

// ConfigFile is the structure of a client configuration file.
type ConfigFile struct {
	Filename string
	Filepath string
	Lang     string
	ConfigKeys
}

// ConfigKeys are the keys in a client configuration file.
type ConfigKeys struct {
	ClientID        string
	ClientSecret    string
	DevToken        string
	RefreshToken    string
	LoginCustomerID string
}

type Comment struct {
	LeftMeta  string
	RightMeta string
}

// Languages defines the idiomatic features of each language in a Google Ads
// API configuration file.
var Languages = map[string]Config{
	"java": {
		Comment: Comment{
			LeftMeta: "#",
		},
		Separator: "=",
		Cfg: ConfigFile{
			Filename: "ads.properties",
			ConfigKeys: ConfigKeys{
				ClientID:        "api.googleads.clientId",
				ClientSecret:    "api.googleads.clientSecret",
				DevToken:        "api.googleads.developerToken",
				RefreshToken:    "api.googleads.refreshToken",
				LoginCustomerID: "api.googleads.loginCustomerId"}}},
	"dotnet": {
		Comment: Comment{
			LeftMeta:  "<!--",
			RightMeta: "-->",
		},
		Cfg: ConfigFile{
			Filename: "App.Config",
			ConfigKeys: ConfigKeys{
				ClientID:        "OAuth2ClientId",
				ClientSecret:    "OAuth2ClientSecret",
				DevToken:        "DeveloperToken",
				RefreshToken:    "OAuth2RefreshToken",
				LoginCustomerID: "LoginCustomerId"}}},
	"php": {
		Comment: Comment{
			LeftMeta: ";",
		},
		Separator: "=",
		Cfg: ConfigFile{
			Filename: "google_ads_php.ini",
			ConfigKeys: ConfigKeys{
				ClientID:        "clientId",
				ClientSecret:    "clientSecret",
				DevToken:        "developerToken",
				RefreshToken:    "refreshToken",
				LoginCustomerID: "loginCustomerId"}}},
	"python": {
		Comment: Comment{
			LeftMeta: "#",
		},
		Separator: ":",
		Cfg: ConfigFile{
			Filename: "google-ads.yaml",
			ConfigKeys: ConfigKeys{
				ClientID:        "client_id",
				ClientSecret:    "client_secret",
				DevToken:        "developer_token",
				RefreshToken:    "refresh_token",
				LoginCustomerID: "login_customer_id"}}},
	"ruby": {
		Comment: Comment{
			LeftMeta: "#",
		},
		Separator: "=",
		Cfg: ConfigFile{
			Filename: "google_ads_config.rb",
			ConfigKeys: ConfigKeys{
				ClientID:        "c.client_id",
				ClientSecret:    "c.client_secret",
				DevToken:        "c.developer_token",
				RefreshToken:    "c.refresh_token",
				LoginCustomerID: "c.login_customer_id"}}}}

// swapMap reverses the keys and values of m.
func swapMap(m map[string]interface{}) map[string]string {
	swapped := make(map[string]string, len(m))
	for k, v := range m {
		swapped[v.(string)] = k
	}
	return swapped
}

// GetConfigKeysInLang returns the key name in the configuration file
// based on the given language. For example, "client_id" is returned with
// "ClientID" for Python.
func (c *ConfigFile) GetConfigKeysInLang(key string) string {
	s := structs.New(Languages[c.Lang].Cfg.ConfigKeys)
	return s.Field(key).Value().(string)
}

// SetConfigKeys updates the value of the given key in ConfigFile.ConfigKeys.
func (c *ConfigFile) SetConfigKeys(k, v string) {
	structs.New(&c.ConfigKeys).Field(k).Set(v)
}

// UpdateConfigKeys updates attributes in ConfigFile.ConfigKeys from keyValue map.
// The keys in keyValue must match the names in ConfigFile.ConfigKeys, else
// they will be ignored.
func (c *ConfigFile) UpdateConfigKeys(keyValue map[string]string) {
	swappedMap := swapMap(structs.Map(Languages[c.Lang].Cfg.ConfigKeys))
	for k, v := range keyValue {
		if mappedK, ok := swappedMap[k]; ok {
			c.SetConfigKeys(mappedK, v)
		}
	}
}

// ReplaceConfigFromReader reads configuration file content from io.Reader
// according to a specific language config file syntax. It inserts the new
// key-value pair and comments out the existing one if found.
func (c *ConfigFile) ReplaceConfigFromReader(key, value string, r io.Reader) string {
	var buf bytes.Buffer

	// Insert the new key-value pair at the "top" of the file. "Top" is
	// the topmost position that is syntactically correct based on the language.
	// And then it finds the line with the old config key and comments it out.
	comment := Languages[c.Lang].Comment
	scanner := bufio.NewScanner(r)
	for i := 0; scanner.Scan(); i++ {
		line := scanner.Text() + "\n"
		trimmedLine := strings.TrimSpace(line)
		langKey := c.GetConfigKeysInLang(key)

		// Found the line with old config key and comment it out
		if !strings.HasPrefix(trimmedLine, comment.LeftMeta) && strings.Contains(trimmedLine, langKey) {
			buf.WriteString(comment.LeftMeta + trimmedLine + comment.RightMeta + "\n")
		} else {
			buf.WriteString(line)
		}

		// Add a line with the new config value
		switch c.Lang {
		case "dotnet":
			if !strings.HasPrefix(trimmedLine, comment.LeftMeta) && strings.Contains(trimmedLine, "<GoogleAdsApi>") {
				buf.WriteString(c.configLineStr(key, value))
			}
		case "php":
			if !strings.HasPrefix(trimmedLine, comment.LeftMeta) {
				if (key == DevToken && strings.Contains(trimmedLine, "[GOOGLE_ADS]")) ||
					strings.Contains(trimmedLine, "[OAUTH2]") {
					buf.WriteString(c.configLineStr(key, value))
				}
			}
		case "ruby":
			if !strings.HasPrefix(trimmedLine, comment.LeftMeta) && strings.Contains(trimmedLine, "Google::Ads::GoogleAds::Config.new") {
				buf.WriteString(c.configLineStr(key, value))
			}
		default:
			if i == 0 {
				buf.WriteString(c.configLineStr(key, value))
			}
		}
	}

	return buf.String()
}

// ReplaceConfig replaces a value in ConfigFile.ConfigKeys and its
// configuration file.
func (c *ConfigFile) ReplaceConfig(key, value string) string {
	c.SetConfigKeys(key, value)

	// Create a temp file
	tmpfile, err := ioutil.TempFile("", "googleadsapi_client_lib_config")
	if err != nil {
		log.Fatalf("ERROR: Problem creating temp file: %s", err)
	}
	defer tmpfile.Close()

	// Open config file
	configFp := filepath.Join(c.Filepath, c.Filename)
	f, err := os.Open(configFp)
	if err != nil {
		log.Fatalf("ERROR: Problem opening config file: %s", err)
	}
	defer f.Close()

	// Replace with new config value and write to temp file
	newConfigStr := c.ReplaceConfigFromReader(key, value, f)
	if _, err := tmpfile.Write([]byte(newConfigStr)); err != nil {
		log.Fatalf("ERROR: Cannot write to temp config file (%s): %s",
			tmpfile.Name(), err)
	}

	f.Close()
	tmpfile.Close()

	// Swap new config file for the old one, and backup the old file
	backupFp := configFp + "_" + time.Now().Format("2006-01-02_15-04-05")
	log.Printf("Backing up config file %s to %s...", configFp, backupFp)
	if err = os.Rename(configFp, backupFp); err != nil {
		log.Fatalf("ERROR: Cannot rename config file from (%s) to (%s): %s",
			configFp, backupFp, err)
	} else {
		log.Printf("Creating a new config file %s...", configFp)
		if err = os.Rename(tmpfile.Name(), configFp); err != nil {
			log.Fatalf("ERROR: Cannot rename config file from (%s) to (%s): %s",
				tmpfile.Name(), configFp, err)
		}
	}

	return backupFp
}

// configLineStr returns a configuration file line formatted for the
// specified language.
func (c *ConfigFile) configLineStr(key, value string) (line string) {
	separator := Languages[c.Lang].Separator
	field := c.GetConfigKeysInLang(key)

	switch strings.ToLower(c.Lang) {
	case "java":
		line = field + separator + value
	case "php":
		line = field + separator + " \"" + value + "\""
	case "ruby":
		line = field + separator + " \"" + value + "\""
	case "python":
		line = field + separator + value
	case "dotnet":
		line = "<add key=\"" + field + "\" value=\"" + value + "\"/>"
	}
	return line + "\n"
}

// ListLanguages returns a slice of supported languages.
func ListLanguages() []string {
	var langs = make([]string, 0)
	for k := range Languages {
		langs = append(langs, k)
	}
	return langs
}

// Contains tests if a string exists in a slice of strings.
func Contains(s []string, str string) bool {
	for _, n := range s {
		if str == n {
			return true
		}
	}
	return false
}

// parseKeyValueLine parses the given line into a key-value pair. The line
// cannot be a comment.
func parseKeyValueLine(c ConfigFile, line string) (string, string, error) {
	separator := Languages[c.Lang].Separator
	if idx := strings.Index(line, separator); idx >= 0 {
		if key := strings.TrimSpace(line[:idx]); len(key) > 0 {
			return key, findFirstValue(line[idx+1:]), nil
		}
	}
	return "", "", fmt.Errorf("Cannot parse key-value pair from this line: %s", line)
}

// findFirstValue returns the first value that contains alphanumeric
// characters potentially with some special characters.
func findFirstValue(k string) string {
	quotedStr := regexp.MustCompile("[\\w\\-\\./_]+")
	matches := quotedStr.FindAllString(k, -1)
	if len(matches) > 0 {
		return matches[0]
	}
	return strings.TrimSpace(k)
}

// ParseKeyValueFile reads a configuration file with keys and values separated
// by a language specific separator, and returns a ConfigFile.
func ParseKeyValueFile(lang, filepath string) (c ConfigFile, err error) {
	keyValue := make(map[string]string, 0)
	c, _ = GetConfigFile(lang, filepath)
	separator := Languages[c.Lang].Separator
	comment := Languages[c.Lang].Comment

	f, err := os.Open(filepath)
	if err != nil {
		return c, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skips comments
		if strings.HasPrefix(line, comment.LeftMeta) {
			continue
		}

		if strings.Contains(line, separator) {
			if k, v, err := parseKeyValueLine(c, line); err != nil {
				log.Print(err)
			} else {
				keyValue[k] = v
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return c, err
	}

	c.UpdateConfigKeys(keyValue)

	return c, nil
}

// ParseXMLFile parses the file content given in filepath and returns
// a ConfigFile struct with the given attributes in the file.
func ParseXMLFile(filepath string) (c ConfigFile, err error) {
	var keyValue = make(map[string]string)
	c, _ = GetConfigFile("dotnet", filepath)

	type Property struct {
		Key   string `xml:"key,attr"`
		Value string `xml:"value,attr"`
	}

	type DotNetXML struct {
		XMLName    xml.Name   `xml:"configuration"`
		Properties []Property `xml:"GoogleAdsApi>add"`
	}

	f, err := os.Open(filepath)
	if err != nil {
		return c, err
	}
	defer f.Close()

	inputBytes, _ := ioutil.ReadAll(f)
	options := DotNetXML{}
	err = xml.Unmarshal([]byte(inputBytes), &options)
	if err != nil {
		return c, err
	}

	for _, prop := range options.Properties {
		keyValue[prop.Key] = prop.Value
	}

	c.UpdateConfigKeys(keyValue)

	return c, nil
}

// IsPII returns true when the given string is PII (peronsal identifiable
// information), else false.
func IsPII(s string) bool {
	return Contains(PIIWords, s)
}

// GetConfigFile returns a ConfigFile containing config filepath and filename.
// When overridePath is an empty string, the function will retrieve the filepath and
// filename from the default location in the file system.
func GetConfigFile(lang, overridePath string) (ConfigFile, error) {
	if overridePath == "" {
		return GetDefaultConfigFile(lang)
	}

	lang = strings.ToLower(lang)
	return ConfigFile{
		Filepath: filepath.Dir(overridePath),
		Filename: filepath.Base(overridePath),
		Lang:     lang}, nil
}

// GetDefaultConfigFile returns the default config path of Google Ads API client
// library.
func GetDefaultConfigFile(lang string) (ConfigFile, error) {
	var cfg ConfigFile

	usr, err := user.Current()
	if err != nil {
		return cfg, err
	}

	if _, ok := Languages[lang]; ok {
		cfg.Filepath = usr.HomeDir
		cfg.Filename = Languages[lang].Cfg.Filename
		cfg.Lang = lang
	}

	return cfg, nil
}

// Print prints out the keys and values in ConfigFile.ConfigKeys.
func (c *ConfigFile) Print(hidePII bool) {
	log.Printf("Config keys and values:")
	keys := reflect.TypeOf(c.ConfigKeys)
	vals := reflect.ValueOf(c.ConfigKeys)
	for i := 0; i < keys.NumField(); i++ {
		k := keys.Field(i).Name
		v := vals.Field(i)
		if hidePII && IsPII(k) && v.String() != "" {
			v = reflect.ValueOf("******************* (hidden)")
		} else if v.String() == "" {
			v = reflect.ValueOf("<empty>")
		}
		log.Printf("\t%s = %s", k, v)
	}
}

// Validate returns true when all the values in ConfigFile.ConfigKeys meet
// the requirements. When it returns false, the returned error includes
// each reason why the attribute fails validation.
func (c *ConfigFile) Validate() (bool, error) {
	valid := true
	var errMsg string
	var err error

	re := regexp.MustCompile("[[:alnum:]_\\-]+")
	if !re.MatchString(c.DevToken) {
		valid = false
		errMsg += fmt.Sprintf("Dev token is invalid. Value: %s\n", c.DevToken)
	}

	if !strings.HasSuffix(c.ClientID, "apps.googleusercontent.com") {
		valid = false
		errMsg += fmt.Sprintf(
			"ClientID does not end with apps.googleusercontent.com. Value: %s\n",
			c.ClientID)
	}

	if strings.Contains(c.LoginCustomerID, "-") {
		valid = false
		errMsg += fmt.Sprintf(
			"LoginCustomerID cannot have dashes. Value: %s\n",
			c.LoginCustomerID)
	}

	keys := reflect.TypeOf(c.ConfigKeys)
	vals := reflect.ValueOf(c.ConfigKeys)
	for i := 0; i < vals.NumField(); i++ {
		k := keys.Field(i).Name
		v := vals.Field(i)

		if Contains(RequiredKeys, k) && v.String() == "" {
			valid = false
			errMsg += fmt.Sprintf("%s is empty.\n", k)
		}

		if strings.Contains(v.String(), "INSERT") {
			valid = false
			errMsg += fmt.Sprintf("%s needs to be updated. Value: %s\n", k, v.String())
		}
	}

	if errMsg != "" {
		err = fmt.Errorf("%s", errMsg)
	}

	return valid, err
}

// MinGoVersion test for the minimum version of Go required.
func MinGoVersion() error {
	return checkGoVersion(runtime.Version())
}

func checkGoVersion(v string) error {
	majorMin := 1
	minorMin := 11

	parts := strings.Split(sanitizeVersion(v), ".")
	if len(parts) < 2 {
		return fmt.Errorf("the given version is too short: %s", v)
	}

	major, err := parseInt(parts[0])
	if err != nil {
		return err
	}

	minor, err := parseInt(parts[1])
	if err != nil {
		return err
	}

	if major <= majorMin && minor < minorMin {
		return fmt.Errorf("minimum required Go version is %d.%d: you are running %s", major, minor, v)
	}
	return nil
}

var semanticVersionRegex = regexp.MustCompile(`^[0-9\.]+`)

func sanitizeVersion(v string) string {
	return semanticVersionRegex.FindString(v)
}

func parseInt(token string) (int, error) {
	num, err := strconv.ParseInt(token, 10, 32)
	if err != nil {
		return -1, fmt.Errorf("could not parse version (%s): %s", token, err)
	}
	return int(num), nil
}
