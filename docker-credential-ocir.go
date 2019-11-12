package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/oracle/oci-go-sdk/common"
	"github.com/oracle/oci-go-sdk/common/auth"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type DockerToken struct {
	Token string `json:"token"`
	Scope string `json:"scope,omitempty"`
	Expires int64 `json:"expires_in,omitempty"`
}

type TokenResponse struct {
	ServerURL string `json:"ServerURL"`
	Username string `json:"Username"`
	Secret string `json:"Secret"`
}

func main() {
	ip, err := auth.InstancePrincipalConfigurationProvider()
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

		if rawUrl[len(rawUrl) - 8:] != ".ocir.io" {
			fmt.Println("Only *.ocir.io registry URLs are supported")
			os.Exit(1)
		}

		tokenUrl := fmt.Sprintf("https://%s/20180419/docker/token", rawUrl)

		registryUrl, err := url.Parse(tokenUrl)
		if err != nil {
			panic(err)
		}

		cl, err := common.NewClientWithConfig(ip)
		if err != nil {
			panic(err)
		}

		cl.Host = registryUrl.Host

		resp, err := cl.Call(context.Background(), &http.Request{
			Method:           "GET",
			URL:              registryUrl,
			Header:           map[string][]string{
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
