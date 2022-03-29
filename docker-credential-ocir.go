package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/oracle/oci-go-sdk/v63/common"
	"github.com/oracle/oci-go-sdk/v63/common/auth"
)

type DockerToken struct {
	Token   string `json:"token"`
	Scope   string `json:"scope,omitempty"`
	Expires int64  `json:"expires_in,omitempty"`
}

type TokenResponse struct {
	ServerURL string `json:"ServerURL"`
	Username  string `json:"Username"`
	Secret    string `json:"Secret"`
}

func envOr(key, def string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return def
}

var (
	usr, _  = user.Current()
	dir     = usr.HomeDir
	ociConf = filepath.Join(dir, ".oci", "config")
)

func main() {
	var cp common.ConfigurationProvider
	var err error
	if filepath.Base(os.Args[0]) == "docker-credential-ocir" {
		cp, err = auth.InstancePrincipalConfigurationProvider()
	} else {
		cp, err = common.ConfigurationProviderFromFileWithProfile(
			envOr("OCI_CLI_CONFIG_FILE", ociConf),
			envOr("OCI_CLI_PROFILE", "DEFAULT"),
			envOr("OCI_CLI_PASSPHRASE", ""), // in an env var?!
		)
	}
	if err != nil {
		panic(err)
	}

	switch strings.ToLower(os.Args[1]) {
	case "store":
		return
	case "erase":
		return
	case "get":
		var rawUrl string
		if _, err := fmt.Scanln(&rawUrl); err != nil {
			panic(err)
		}
		rawUrl = strings.TrimSpace(strings.ToLower(rawUrl))
		if rawUrl[:8] == "https://" {
			rawUrl = rawUrl[8:]
		}

		//		if rawUrl[len(rawUrl)-8:] != ".ocir.io" {
		//			fmt.Println("Only *.ocir.io registry URLs are supported")
		//			os.Exit(1)
		//		}

		// Try to get an example repo manifest
		var realm string
		if resp, err := http.Get(fmt.Sprintf("https://%s/v2/", rawUrl)); err != nil {
			panic(err)
		} else {
			resp.Body.Close()
			authHeader := resp.Header.Get("www-authenticate")
			// Bearer realm="https://phx.ocir.io/20180419/docker/token",service="phx.ocir.io",scope=""
			parts := strings.Split(authHeader, " ")

			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				panic("Unexpected www-authenticate header")
			}
			values := map[string]string{}
			for _, part := range strings.Split(parts[1], ",") {
				kv := strings.SplitN(part, "=", 2)
				if len(kv) == 2 {
					values[strings.ToLower(kv[0])] = strings.Trim(kv[1], "\"")
				}
			}
			realm = values["realm"]
		}

		if realm == "" {
			panic("no realm returned in initial registry handshake")
		}

		registryUrl, err := url.Parse(realm)
		if err != nil {
			panic(err)
		}

		cl, err := common.NewClientWithConfig(cp)
		if err != nil {
			panic(err)
		}

		cl.Host = registryUrl.Host

		resp, err := cl.Call(context.Background(), &http.Request{
			Method: "GET",
			URL:    registryUrl,
			Header: map[string][]string{
				"Accept": {"application/json"},
			},
		})
		if err != nil {
			panic(err)
		}
		if resp.Body != nil {
			defer resp.Body.Close()
		} else {
			panic("no body in return")
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}
		var token DockerToken
		if err := json.Unmarshal(body, &token); err != nil {
			panic(err)
		}

		tr := TokenResponse{
			ServerURL: rawUrl,
			Username:  "BEARER_TOKEN",
			Secret:    token.Token,
		}

		out, err := json.Marshal(tr)
		if err != nil {
			panic(err)
		}

		fmt.Println(string(out))
	}
}
