package oauth

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/googleads/google-ads-doctor/oauthdoctor/diag"
)

// privateKey is a fake value to bypass x509.ParsePKCS1PrivateKey()
// parsing error.
const fakePrivateKey = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDHxZaeBzu7JOqY
HV6PKCMb+tsiFRsFiY1kGyexJXzLYy39sRL6Xsg1jxMUx6TdHRzSP06N225DbUs3
D1vbEZAp91qdJouvK9Mbf+1eQ5d4yeDQZMqdJJoBbmckCGcR/QhXwZ2wn0Gw0AhH
Af4EsCX3d8Cxtu5zxzR0GIE6YUONt3RZdBgpddQzlFCRGZBkz+c7jUejT7iBdwix
DaW7itqYO5TETaqwF1bzefD1BSrIjATfdcyUdpMtMTNKfVTZ9Eh3eKy+qIBdTtH/
avvf4AVc2yN1K/+occ7doEclSd3W7vTi51w3VT8TdA9aFVz0RbxP3Y2px0zWX12f
unLeBSWbAgMBAAECggEARfVZagzha4ehidSbJSnqpaVDMRvQAy/o7loeG8ije7xH
QlTM7xXbKfppNbk2cGJ+EdiuoznpUr6G/QipY72yTSf8uRTjDNydiL9TelPUSy3z
RzdMxxwmvIKTpwg0RBXm4oiAtvYGdKtdgrRdZvniyddLiVClD7F+mntsYevm0s0C
AgOFwfz/pGun7EIKGaDCo418u9nHbhQUyscfqmZEx9lTNRwcGAbxoi4EC0OVifDr
k+yoFwXjJG1mcMxZUaJTJYwP65rCUSviMUvzd7BGfenNLZTxSiHEZ/hHa3glLfVy
ggf1WA+74Mu/0KBV4Ya6ZAkHHw6hVET+8+faG2h1mQKBgQDo11lW7k3nflHCS9qD
5FUsZeesE+1sQjCCpA7TvwWnJ3avYP8R1dyHiwPXH5kMl9YKn7N/DxOqzsvFC0Qh
1UBJQjWlm2YPK5I3FYX66lwkKNibaU3sq+Ly1CFCGqrKZa839otd5sVOol30uplF
018G8aTiqhHuyc1fjyx+8jSn8wKBgQDbpDhT6W+HdFrDO5TVae6pGHMO4GMrmvQm
R8IHeMDuU4bN4XL1Uj1skp+apcqcAareFItIi03VyZnf7koGQujqgFLEZK2naUXA
IOD2gAz0YZEiZbFeXKLdVX1DnAvz6cbhRp2qpR+Ky+vOHEjV6mY+YuKDz/bzgMSe
tkc47NnduQKBgQDaX8ZpcnzcJSvW9z9MvaRoTHbYe6QMCZPnoqhJTXmmyKtWVrlC
5/m5odaLNxZaqjjTo+47t08xvlt8RVG0DYYKby9TT4iLp8itIuGSb6TVQP3N3Bh6
ZMcoCW3bypjt1CpeaTtSaTIZyswlz7Aavd/86jtDXlANTXTxL52CvfRGowKBgQCW
RGr5Fbrk/DjgWxH/VEMg0wZcxi1y9sdUrUFU5Utxghm3HygMKKC3eDTTk9vjEcz5
tSp5jjzJJ+0rZBam4/3/+Z0mmg6oe4Bp6tSeMIssYtftpY9MlKokLUnPCKKw1F7p
XuudhOzog40nbPhzybL7uaFpNs2oWI+sWd6uVnTTmQKBgF450f5zWvUFv/00WUY+
1YWZtYDWV/tzwauz4+7pNC4MNGQQpOFEVEucPVfKMX/vbm1h1DJTVPzBgdkw7nfk
ckD3dVK4nva0qzdpfKOX/QHX/21Y0ivELswnmZ488IYrwWJpJta0ZvD1/SpvAgJb
+VLBOI3lpGHlxgh2ZTcanfsV
-----END PRIVATE KEY-----`

func TestSimulateServiceAccFlow(t *testing.T) {
	s, close := setupFakeOAuthServer()
	tokenURL = s.URL + "/token" // overriding the code to use this fake tokenURL
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
			desc: "Test: OAuth succeeds",
			c: Config{
				ConfigFile: diag.ConfigFile{
					ServiceAccountInfo: diag.ServiceAccountInfo{
						PrivateKey: fakePrivateKey,
					},
				},
			},
			ts: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte(`{"resourceName": "customers/1234567890", "id": "1234567890"}`))
			})),
			want: "OAuth test passed",
		},
		{
			desc: "Test: OAuth fails",
			c: Config{
				ConfigFile: diag.ConfigFile{
					ServiceAccountInfo: diag.ServiceAccountInfo{
						PrivateKey: fakePrivateKey,
					},
				},
			},
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

		tt.c.simulateServiceAccFlow()

		if !strings.Contains(got.String(), tt.want) {
			t.Errorf("\n[%s]\ngot: %s\nwant substring: %s", tt.desc, got.String(), tt.want)
		}
	}
}
