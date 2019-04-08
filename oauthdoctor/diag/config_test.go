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
package diag_test

import (
	"log"
	"oauthdoctor/diag"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestValidate(t *testing.T) {
	const goodDevToken = "_asdfbasd_-0adsfaw8762"
	const goodClientID = "012345678-8hafs7yfas0f0fh.apps.googleusercontent.com"
	const goodToken = "89yashfoasuf0ujafi0f"
	const goodSecret = "09aufj0aj0ufa8s"

	tests := []struct {
		cfg    diag.ConfigFile
		want   bool
		errstr string
	}{
		{
			cfg: diag.ConfigFile{
				ConfigKeys: diag.ConfigKeys{
					DevToken:        goodDevToken,
					ClientID:        goodClientID,
					ClientSecret:    goodSecret,
					RefreshToken:    goodToken,
					LoginCustomerID: "1111111111",
				},
			},
			want:   true,
			errstr: "nil",
		}, // Everything passes
		{
			cfg: diag.ConfigFile{
				ConfigKeys: diag.ConfigKeys{
					DevToken:     "INSERT_DEV_TOKEN_HERE",
					ClientID:     goodClientID,
					ClientSecret: goodSecret,
					RefreshToken: goodToken,
				},
			},
			want:   false,
			errstr: "DevToken",
		}, // Invalid DevToken
		{
			cfg: diag.ConfigFile{
				ConfigKeys: diag.ConfigKeys{
					DevToken:     goodDevToken,
					ClientID:     "randomClientID",
					ClientSecret: goodSecret,
					RefreshToken: goodToken,
				},
			},
			want:   false,
			errstr: "ClientID",
		}, // Invalid ClientID
		{
			cfg: diag.ConfigFile{
				ConfigKeys: diag.ConfigKeys{
					DevToken:     goodDevToken,
					ClientID:     goodClientID,
					RefreshToken: goodToken,
				},
			},
			want:   false,
			errstr: "ClientSecret",
		}, // Missing a required key
		{
			cfg: diag.ConfigFile{
				ConfigKeys: diag.ConfigKeys{
					LoginCustomerID: "111-111-1111",
				},
			},
			want:   false,
			errstr: "LoginCustomerID",
		}, // LoginCustomerID cannot have dashes
	}

	for _, test := range tests {
		got, err := test.cfg.Validate()
		if got != test.want || !strings.Contains(errstring(err), test.errstr) {
			t.Errorf("Wrong result - got: %+v, want: %+v, got err: %s, but missing %s in error msg",
				got, test.want, errstring(err), test.errstr)
		}
	}
}

//TODO: add tests
func TestReplaceConfig(t *testing.T) {
	tests := []struct {
		key   string
		value string
		cfg   diag.ConfigFile
		input string
		want  string
	}{
		{
			key:   diag.RefreshToken,
			value: "newValue",
			cfg:   diag.ConfigFile{Lang: "python"},
			input: `developer_token: GoodDevToken
client_secret: GoodClientSecret
refresh_token: GoodRefreshToken`,
			want: `developer_token: GoodDevToken
refresh_token:newValue
client_secret: GoodClientSecret
#refresh_token: GoodRefreshToken
`,
		},
	}

	for _, test := range tests {
		got := test.cfg.ReplaceConfigFromReader(test.key, test.value, strings.NewReader(test.input))

		if got != test.want {
			t.Errorf("Wrong result - got: %s, want: %s", got, test.want)
		}
	}
}

func TestParseKeyValueFile(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current dir: %s", err)
	}

	tests := []struct {
		configPath string
		lang       string
		want       diag.ConfigFile
	}{
		{
			configPath: filepath.Join(dir, "testdata", "config_file1"),
			lang:       "python",
			want: diag.ConfigFile{
				Filepath: filepath.Join(dir, "testdata"),
				Filename: "config_file1",
				Lang:     "python",
				ConfigKeys: diag.ConfigKeys{
					ClientID:     "0123456789-GoodClientID.apps.googleusercontent.com",
					ClientSecret: "GoodClientSecret",
					DevToken:     "GoodDevToken",
					RefreshToken: "1/PG1Ap6P-Good_Refresh_Token",
				},
			},
		}, // Python
		{
			configPath: filepath.Join(dir, "testdata", "config_file2"),
			lang:       "ruby",
			want: diag.ConfigFile{
				Filepath: filepath.Join(dir, "testdata"),
				Filename: "config_file2",
				Lang:     "ruby",
				ConfigKeys: diag.ConfigKeys{
					ClientID: "GoodClientID",
				},
			},
		}, // Ruby: Missing required config keys with comments
		{
			configPath: filepath.Join(dir, "testdata", "config_file3"),
			lang:       "php",
			want: diag.ConfigFile{
				Filepath: filepath.Join(dir, "testdata"),
				Filename: "config_file3",
				Lang:     "php",
				ConfigKeys: diag.ConfigKeys{
					ClientID:     "GoodClientID",
					ClientSecret: "GoodClientSecret",
					DevToken:     "GoodDevToken",
					RefreshToken: "GoodRefreshToken",
				},
			},
		}, // PHP: Can parse values with quotes
		{
			configPath: filepath.Join(dir, "testdata", "config_file4"),
			lang:       "java",
			want: diag.ConfigFile{
				Filepath: filepath.Join(dir, "testdata"),
				Filename: "config_file4",
				Lang:     "java",
				ConfigKeys: diag.ConfigKeys{
					ClientID:     "GoodClientID",
					ClientSecret: "GoodClientSecret",
					DevToken:     "GoodDevToken",
					RefreshToken: "GoodRefreshToken",
				},
			},
		}, // Java
	}

	for _, test := range tests {
		got, err := diag.ParseKeyValueFile(test.lang, test.configPath)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("KevValueFile mismatch - got: %+v, want: %+v, err: %s",
				got, test.want, errstring(err))
		}
	}
}

func TestParseXMLFile(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current dir: %s", err)
	}

	tests := []struct {
		configPath string
		lang       string
		want       diag.ConfigFile
	}{
		{
			configPath: filepath.Join(dir, "testdata", "xml_config_file1"),
			lang:       "dotnet",
			want: diag.ConfigFile{
				Filepath: filepath.Join(dir, "testdata"),
				Filename: "xml_config_file1",
				Lang:     "dotnet",
				ConfigKeys: diag.ConfigKeys{
					ClientID:     "0123456789-GoodClientID.apps.googleusercontent.com",
					ClientSecret: "GoodClientSecret",
					DevToken:     "GoodDevToken",
					RefreshToken: "1/PG1Ap6P-Good_Refresh_Token",
				},
			},
		}, // Can parse DotNet XML with sp
	}

	for _, test := range tests {
		got, err := diag.ParseXMLFile(test.configPath)
		if !reflect.DeepEqual(got, test.want) {
			t.Errorf("KevValueFile mismatch - got: %+v, want: %+v, err: %s",
				got, test.want, errstring(err))
		}
	}
}

func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return "nil"
}
