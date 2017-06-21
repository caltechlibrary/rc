// apiexplorer is a demo program showing how to use and restclient package
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	// Caltech Library packages
	"github.com/caltechlibrary/cli"
	"github.com/caltechlibrary/rc"

	// 3rd Party Packages
	x2j "github.com/basgys/goxml2json"
)

var (
	usage = `USAGE %s [OPTIONS] URL`

	description = `%s is a demo program exercising the rc Golang package.`

	examples = ``

	// Standard Options
	showHelp    bool
	showLicense bool
	showVersion bool
	outputFName string

	// Application Options
	authMethod string
	method     string
	userName   string
	userSecret string
	payload    string
	asJSON     bool
)

func init() {
	// Standard Options
	flag.BoolVar(&showHelp, "h", false, "display help")
	flag.BoolVar(&showHelp, "help", false, "display help")
	flag.BoolVar(&showLicense, "l", false, "display license")
	flag.BoolVar(&showLicense, "license", false, "display license")
	flag.BoolVar(&showVersion, "v", false, "display version")
	flag.BoolVar(&showVersion, "version", false, "display version")
	flag.StringVar(&outputFName, "o", "", "output filename")
	flag.StringVar(&outputFName, "output", "", "output filename")

	// Application Options
	flag.StringVar(&authMethod, "auth", "", "set authorization type (e.g. oauth, shib)")
	flag.StringVar(&userName, "u", "", "set username for authentication")
	flag.StringVar(&userName, "username", "", "set username for authentication")
	flag.StringVar(&userSecret, "p", "", "set user secret to use for authentication")
	flag.StringVar(&userSecret, "password", "", "set user secret to use for authentication")
	flag.StringVar(&method, "method", "GET", "set the http method to use for request")
	flag.StringVar(&payload, "payload", "", "A JSON structure holding the payload data")
	flag.BoolVar(&asJSON, "as-json", false, "Convert XML to JSON before output")
}

func main() {
	appName := path.Base(os.Args[0])

	// Configuration and command line interation
	cfg := cli.New(appName, appName, fmt.Sprintf(rc.LicenseText, appName, rc.Version), rc.Version)
	cfg.UsageText = fmt.Sprintf(usage, appName)
	cfg.DescriptionText = fmt.Sprintf(description, appName)
	cfg.ExampleText = examples

	userName = cfg.CheckOption("username", cfg.MergeEnv("username", userName), false)
	userSecret = cfg.CheckOption("password", cfg.MergeEnv("password", userSecret), false)
	authMethod = cfg.CheckOption("auth_method", cfg.MergeEnv("auth_method", authMethod), false)

	flag.Parse()
	args := flag.Args()

	if showHelp == true {
		fmt.Println(cfg.Usage())
		os.Exit(0)
	}

	if showLicense == true {
		fmt.Println(cfg.License())
		os.Exit(0)
	}

	if showVersion == true {
		fmt.Println(cfg.Version())
		os.Exit(0)
	}

	out, err := cli.Create(outputFName, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
	defer cli.CloseFile(outputFName, out)

	authType := rc.AuthNone
	switch strings.TrimSpace(strings.ToLower(authMethod)) {
	case "basic":
		authType = rc.BasicAuth
	case "oath":
		authType = rc.OAuth
	case "shib":
		authType = rc.Shibboleth
	}

	data := map[string]string{}
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &data); err != nil {
			log.Fatal(err)
		}
	}

	for _, arg := range args {
		u, err := url.Parse(arg)
		if err != nil {
			log.Fatal(err)
		}
		api, err := rc.New(arg, authType, userName, userSecret)
		if err != nil {
			log.Fatal(err)
		}
		if src, err := api.Request(method, u.Path, data); err == nil {
			if asJSON == true {
				// FIXME: only do xml Unmarshal if response is XML
				if s, err := x2j.Convert(bytes.NewReader(src)); err != nil {
					log.Fatal(err)
				} else {
					fmt.Fprintf(out, "%s\n", s)
					os.Exit(0)
				}
			}
			fmt.Fprintf(out, "%s\n", src)
			os.Exit(0)
		} else {
			log.Fatal(err)
		}
	}
}
