package oauth

import (
	"io/ioutil"
	"log"
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

func disableOutput(t *testing.T) func() {
	log.SetOutput(ioutil.Discard)

	var err error
	stdout := os.Stdout
	os.Stdout, err = os.Create(os.DevNull)
	if err != nil {
		t.Fatalf("Unable to create /dev/null: %s", err)
	}
	stdin := readStdin

	enableOutput := func() {
		os.Stdout = stdout
		readStdin = stdin
	}

	return enableOutput
}

func TestReplaceDevToken(t *testing.T) {
	enableOutput := disableOutput(t)
	defer enableOutput()

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
	enableOutput := disableOutput(t)
	defer enableOutput()

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
