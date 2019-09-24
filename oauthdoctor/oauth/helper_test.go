package oauth

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/googleads/google-ads-doctor/oauthdoctor/diag"
)

type FakeConfig struct {
	cfgFile diag.ConfigFile
}

func (c *FakeConfig) ReplaceConfig(k, v string) string {
	c.cfgFile.SetConfigKeys(k, v)
	return ""
}

func TestReplaceCloudCredentials(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	originalReadID := getClientID
	originalReadSecret := getClientSecret
	defer func() {
		getClientID = originalReadID
		getClientSecret = originalReadSecret
	}()

	test := struct {
		desc         string
		c            FakeConfig
		clientID     string
		clientSecret string
	}{
		desc: "Successful Replacement",
		c: FakeConfig{
			cfgFile: diag.ConfigFile{
				ConfigKeys: diag.ConfigKeys{
					ClientID:     "oldID",
					ClientSecret: "oldSecret",
				},
			},
		},
		clientID:     "newID",
		clientSecret: "newSecret",
	}

	getClientID = func() string {
		return test.clientID
	}
	getClientSecret = func() string {
		return test.clientSecret
	}

	replaceCloudCredentials(&test.c)

	if test.c.cfgFile.ClientID != test.clientID || test.c.cfgFile.ClientSecret != test.clientSecret {
		t.Errorf("[%s] got: (ClientID=%s, ClientSecret=%s), want: (ClientID=%s, ClientSecret=%s)",
			test.desc, test.c.cfgFile.ClientID, test.c.cfgFile.ClientSecret, test.clientID, test.clientSecret)
	}
}

func disableStdio(t *testing.T) func() {
	log.SetOutput(ioutil.Discard)

	var err error
	stdout := os.Stdout
	os.Stdout, err = os.Create(os.DevNull)
	if err != nil {
		t.Fatalf("Unable to create /dev/null: %s", err)
	}
	stdin := readStdin

	enableStdio := func() {
		os.Stdout = stdout
		readStdin = stdin
	}

	return enableStdio
}

func TestReplaceDevToken(t *testing.T) {
	enableStdio := disableStdio(t)
	defer enableStdio()

	test := struct {
		desc string
		c    FakeConfig
		want string
	}{
		desc: "Successful replacement",
		c: FakeConfig{
			cfgFile: diag.ConfigFile{
				ConfigKeys: diag.ConfigKeys{
					DevToken: "oldDevToken",
				},
			},
		},
		want: "newDevToken",
	}

	readStdin = func() string {
		return test.want
	}

	replaceDevToken(&test.c)

	if test.want != test.c.cfgFile.DevToken {
		t.Errorf("[%s] got: %s, want: %s", test.desc, test.c.cfgFile.DevToken, test.want)
	}
}

func TestReplaceRefreshToken(t *testing.T) {
	enableStdio := disableStdio(t)
	defer enableStdio()

	tests := []struct {
		desc  string
		c     FakeConfig
		stdin string
		input string
		want  string
	}{
		{
			desc: "Successful replacement",
			c: FakeConfig{
				cfgFile: diag.ConfigFile{
					ConfigKeys: diag.ConfigKeys{
						RefreshToken: "oldRefreshToken",
					},
				},
			},
			stdin: "Y",
			input: "newRefreshToken",
			want:  "newRefreshToken",
		},
		{
			desc: "No replacement",
			c: FakeConfig{
				cfgFile: diag.ConfigFile{
					ConfigKeys: diag.ConfigKeys{
						RefreshToken: "oldRefreshToken",
					},
				},
			},
			stdin: "N",
			input: "newRefreshToken",
			want:  "oldRefreshToken",
		},
	}

	for _, test := range tests {
		readStdin = func() string {
			return test.stdin
		}

		replaceRefreshToken(&test.c, test.input)

		if test.c.cfgFile.RefreshToken != test.want {
			t.Errorf("[%s] got: %s, want: %s", test.desc, test.c.cfgFile.RefreshToken, test.want)
		}
	}
}

func TestGetAccount(t *testing.T) {
	tests := []struct {
		desc string
		c    Config
		ts   *httptest.Server
		want string
	}{
		{
			desc: "developer-token is in HTTP header",
			c: Config{
				ConfigFile: diag.ConfigFile{
					ConfigKeys: diag.ConfigKeys{
						DevToken: "devToken",
					},
				},
			},
			ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(r.Header["Developer-Token"][0]))
			})),
			want: "devToken",
		},
		{
			desc: "login-customer-id is in HTTP header",
			c: Config{
				ConfigFile: diag.ConfigFile{
					ConfigKeys: diag.ConfigKeys{
						LoginCustomerID: "loginID",
					},
				},
			},
			ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(r.Header["Login-Customer-Id"][0]))
			})),
			want: "loginID",
		},
		{
			desc: "Account info (JSON) is returned",
			c:    Config{},
			ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"resourceName": "customers/1234567890", "id": "1234567890"}`))
			})),
			want: `{"resourceName": "customers/1234567890", "id": "1234567890"}`,
		},
		{
			desc: "Error (JSON) is returned",
			c:    Config{},
			ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"error": "This is an error"}`))
			})),
			want: "This is an error",
		},
	}

	for _, tt := range tests {
		apiURL = tt.ts.URL
		defer tt.ts.Close()

		buf, err := tt.c.getAccount(tt.ts.Client())
		if err != nil && errstring(err) != tt.want {
			t.Errorf("[%s] got: %s, want: %s", tt.desc, errstring(err), tt.want)
		}

		if buf != nil && buf.String() != tt.want {
			t.Errorf("[%s] got: %s, want: %s", tt.desc, buf.String(), tt.want)
		}
	}
}

func TestReadCustomerID(t *testing.T) {
	log.SetOutput(ioutil.Discard)

	original := readStdin
	defer func() { readStdin = original }()

	tests := []struct {
		desc  string
		stdin string
		want  string
	}{
		{
			desc:  "Return a valid customer ID",
			stdin: "123-456-7890",
			want:  "1234567890",
		},
		{
			desc:  "Return original string",
			stdin: "abc",
			want:  "abc",
		},
	}

	for _, test := range tests {
		readStdin = func() string {
			return test.stdin
		}

		got := ReadCustomerID()
		if got != test.want {
			t.Errorf("[%s] got: %s, want: %s\n", test.desc, got, test.want)
		}
	}
}

func errstring(err error) string {
	if err != nil {
		return err.Error()
	}
	return "nil"
}
