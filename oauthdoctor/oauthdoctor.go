// Copyright 2019 Google LLC
//
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
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/googleads/google-ads-doctor/oauthdoctor/diag"
	"github.com/googleads/google-ads-doctor/oauthdoctor/oauth"
)

var (
	oauthTypes = []string{diag.InstalledApp, diag.Web, diag.ServiceAccount}
	language   = flag.String("language", "", "Required: The programming language of Google Ads API client library")
	oauthType  = flag.String("oauthtype", "Required: The OAuth2 type for Google Ads API.", fmt.Sprintf("Values: %s", strings.Join(oauthTypes, ", ")))
	configPath = flag.String("configpath", "", "Optional: An absolute file path for Google Ads API configuration file")
	customerId = flag.String("customerid", "", "Optional: A customer ID. Providing this value avoids prompting for a customer ID during execution.")
	hidePII    = flag.Bool("hidepii", true, "Optional: Suppress output of Personally Identifiable Information")
	sysinfo    = flag.Bool("sysinfo", false, "Optional: Print system information.")
	verbose    = flag.Bool("verbose", false, "Optional: Print out debugging info, such as JSON response")
)

func main() {
	log.SetOutput(os.Stdout)

	if err := diag.MinGoVersion(); err != nil {
		log.Fatal(err)
	}

	flag.Parse()

	if flag.NFlag() < 2 {
		log.Fatalf("Please provide --language and --oauthtype")
	}

	language := strings.ToLower(*language)
	languages := diag.ListLanguages()
	if ok := diag.Contains(languages, language); !ok {
		l := strings.Join(languages, ",")
		log.Fatalf("You specified %s. Supported languages are %s\n", language, l)
	}
	log.Printf("Client library language: %s\n", language)

	// Print system info
	if *sysinfo {
		s := diag.SysInfo{}
		s.Init()
		s.Print()
		diag.PrintIPv4(s.Host)

		err := diag.ConnEndpoint()
		if err != nil {
			log.Printf("Connect to endpoint error: %s", err)
		} else {
			fmt.Printf("Connected to %s\n", diag.ENDPOINT)
		}
	}

	// Verify the existence of the config file
	cfg := diag.GetConfigFile(language, *configPath)
	*configPath = cfg.GetFilepath()
	if _, err := os.Stat(*configPath); os.IsNotExist(err) {
		log.Fatalf("Cannot find config file (%s): %s\n", *configPath, err)
	}
	log.Printf("Google Ads API client library config file: %s\n", *configPath)

	// Verify OAuth type
	if ok := diag.Contains(oauthTypes, *oauthType); !ok {
		log.Fatalf("OAuth type not supported: %s", *oauthType)
	}

	var err error
	// Parse config file and get a map of key:value
	switch language {
	case "dotnet":
		cfg, err = diag.ParseXMLFile(*configPath, *oauthType)
	default:
		cfg, err = diag.ParseKeyValueFile(language, *configPath, *oauthType)
	}
	if err != nil {
		log.Fatalf("Cannot parse %s: %s", *configPath, err)
	}

	cfg.Print(*hidePII)

	if ok, err := cfg.Validate(); !ok {
		log.Printf("Config file validation failed: %s\n", err)
	}

	var cid string
	if strings.TrimSpace(*customerId) == "" {
		cid = oauth.ReadCustomerID()
	} else {
		cid = strings.ReplaceAll(*customerId, "-", "")
	}

	c := oauth.Config{
		ConfigFile: cfg,
		CustomerID: cid,
		OAuthType:  *oauthType,
		Verbose:    *verbose,
	}
	c.SimulateOAuthFlow()
}
