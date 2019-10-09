package oauth

import (
	"log"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
)

var tokenURL = google.JWTTokenURL

func (c *Config) simulateServiceAccFlow() {
	conf := &jwt.Config{
		Email:      c.ConfigFile.ClientEmail,
		PrivateKey: []byte(c.ConfigFile.PrivateKey),
		Scopes:     []string{GoogleAdsApiScope},
		TokenURL:   tokenURL,
		Subject:    c.ConfigFile.DelegatedAccount,
	}
	client := conf.Client(oauth2.NoContext)

	accountInfo, err := c.getAccount(client)
	if err == nil {
		if c.Verbose {
			log.Print(accountInfo.String())
		}
		log.Println("SUCCESS: OAuth test passed with given config file settings.")
	} else {
		c.diagnose(err)
		if c.Verbose {
			log.Println(err)
		}
		log.Println("ERROR: OAuth test failed.")
	}
}
