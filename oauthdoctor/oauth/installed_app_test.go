package oauth

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/googleads/google-ads-doctor/oauthdoctor/diag"
	"golang.org/x/oauth2"
)

func setupFakeOAuthServer() (*httptest.Server, func()) {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		// Should return authorization code back to the user
		w.Write([]byte("code=mockauthcode"))
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		// Should return acccess token back to the user
		w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
		w.Write([]byte("access_token=mocktoken&scope=user&token_type=bearer"))
	})

	server := httptest.NewServer(mux)

	// overriding the endpoint for OAuth2 library
	oauthEndpoint = oauth2.Endpoint{
		AuthURL:  server.URL + "/auth",
		TokenURL: server.URL + "/token",
	}

	return server, func() {
		server.Close()
	}
}

func TestAppFlow(t *testing.T) {
	_, close := setupFakeOAuthServer()
	defer close()

	enableStdio := disableStdio(t)
	defer enableStdio()

	tests := []struct {
		desc string
		c    Config
		ts   *httptest.Server
		want string
	}{
		{
			desc: "OAuth succeeds",
			c: Config{
				ConfigFile: diag.ConfigFile{
					ConfigKeys: diag.ConfigKeys{
						RefreshToken: "fakeToken",
					},
				},
			},
			ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"resourceName": "customers/1234567890", "id": "1234567890"}`))
			})),
			want: "OAuth test passed",
		},
		{
			desc: "OAuth retry succeeds",
			c:    Config{},
			ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"resourceName": "customers/1234567890", "id": "1234567890"}`))
			})),
			want: "OAuth test passed",
		},
		{
			desc: "OAuth fails",
			c:    Config{},
			ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			})),
			want: "OAuth test failed",
		},
	}

	for _, tt := range tests {
		apiURL = tt.ts.URL
		defer tt.ts.Close()

		var got strings.Builder
		log.SetOutput(&got)

		tt.c.simulateAppFlow()

		if !strings.Contains(got.String(), tt.want) {
			t.Errorf("[%s] got: %s, want: %s", tt.desc, got.String(), tt.want)
		}
	}
}

func TestGenAuthCode(t *testing.T) {
	var tests = []struct {
		desc  string
		input string
		want  string
	}{
		{
			desc:  "Windows",
			input: "windows",
			want:  "You are running Windows",
		},
		{
			desc:  "Linux",
			input: "linux",
			want:  "Copy",
		},
	}

	for _, tt := range tests {
		got := genAuthCodePrompt(tt.input)
		if !strings.HasPrefix(got, tt.want) {
			t.Errorf("\n%s\nprompt does not match operating system got=%s\nwant=%s\ninput=%s\n",
				tt.desc, got, tt.want, tt.input)
		}
	}
}
