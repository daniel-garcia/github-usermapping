package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var endpoint = "https://api.github.com/graphql"

func main() {

	cursor := ""
	i := 0
	users := make(map[string]string)
	for {
		response := getList(cursor)
		i = i + 1
		pageInfo := response.Data.Organization.SamlIdentityProvider.ExternalIdentities.PageInfo
		for _, edge := range response.Data.Organization.SamlIdentityProvider.ExternalIdentities.Edges {
			if edge.Node.SamlIdentity.NameID == "" {
				continue
			}
			users[edge.Node.SamlIdentity.NameID] = edge.Node.User.Login
		}
		if pageInfo.HasNextPage == false {
			b, _ := json.MarshalIndent(users, " ", " ")
			fmt.Printf("%s\n", string(b))
			return
		}
		cursor = pageInfo.EndCursor
	}
}

func getList(cursor string) Response {
	github_user := os.Getenv("GITHUB_USER")
	github_token := os.Getenv("GITHUB_TOKEN")
	if len(github_token) == 0 {
		log.Fatal("need GITHUB_TOKEN set")
	}
	client := http.Client{
		Timeout: time.Second * 5,
	}

	query := strings.NewReader(getQuery(os.Getenv("GITHUB_ORG"), cursor))

	req, err := http.NewRequest(http.MethodPost, endpoint, query)
	if err != nil {
		log.Fatalf("could not create request: %s", err)
	}
	req.SetBasicAuth(github_user, github_token)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		log.Fatalf("could not process request: %s", err)
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	response := Response{}
	if err := json.Unmarshal(body, &response); err != nil {
		log.Fatalf("could not unmarshal response: %s", err)
	}
	return response
}

type Response struct {
	Data struct {
		Organization struct {
			SamlIdentityProvider struct {
				ExternalIdentities struct {
					Edges []struct {
						Cursor string `json:"cursor"`
						Node   struct {
							SamlIdentity struct {
								NameID string `json:"nameId"`
							} `json:"samlIdentity"`
							User struct {
								Login string `json:"login"`
							} `json:"user"`
						} `json:"node"`
					} `json:"edges"`
					PageInfo struct {
						EndCursor   string `json:"endCursor"`
						HasNextPage bool   `json:"hasNextPage"`
						StartCursor string `json:"startCursor"`
					} `json:"pageInfo"`
				} `json:"externalIdentities"`
			} `json:"samlIdentityProvider"`
		} `json:"organization"`
	} `json:"data"`
}

const queryTemplate = `{ "query": "{ organization(login: \"%s\") { samlIdentityProvider { externalIdentities(first: 100%s) { pageInfo { endCursor  startCursor  hasNextPage } edges { cursor  node { samlIdentity { nameId  } user { login } } } } } } }" }`

func getQuery(repo, cursor string) string {
	cursorStr := ""
	if len(cursor) > 0 {
		cursorStr = fmt.Sprintf(` after: \"%s\"`, cursor)
	}
	return fmt.Sprintf(queryTemplate, repo, cursorStr)
}
