package git

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type GitHubProvider struct {
	Hostname string
}

var GitHub = GitHubProvider{
	Hostname: "github.com",
}

// GitHub SSH key API
// https://docs.github.com/en/rest/users/keys?apiVersion=2022-11-28

type AddSSHKeyResponseGitHub struct {
	Key string `json:"key"`
	Id int `json:"id"`
	URL string `json:"url"`
	Title string `json:"title"`
	CreatedAt string `json:"created_at"`
	Verified bool `json:"verified"`
	ReadOnly bool `json:"read_only"`
}

func (provider GitHubProvider) AddSSHKey(publicKeyBytes []byte, title string, accessToken string) (int, error) {
	var keyId int
	var err error

	url := "https://api.github.com/user/keys"
	contentType := "application/json"

	// Create request body and encode as buffer
	publicKeyStr := string(publicKeyBytes)
	// Must strip key of newline characters for correct body encoding
	publicKeyStr = strings.Replace(publicKeyStr, "\n", "", -1)
	bodyStr := fmt.Sprintf(`{"title": "%s", "key": "%s"}`, title, publicKeyStr)
 	bodyBuffer := bytes.NewBuffer([]byte(bodyStr))

	// Construct POST request
	// From: https://golangnote.com/request/sending-post-request-in-golang-with-header
	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bodyBuffer)
	if err != nil {
		return keyId, err
	}

	// Format access token header
	accessTokenHeader := fmt.Sprintf("Bearer %s", accessToken)

	// Add request headers
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", accessTokenHeader)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	// Execute add SSH key request
	res, err := client.Do(req)
	if err != nil {
		return keyId, err
	}
	defer res.Body.Close()

	// Read response body -> Get key id
	// From: https://stackoverflow.com/a/31129967
	responseBody := &AddSSHKeyResponseGitHub{}
	err = json.NewDecoder(res.Body).Decode(responseBody)
	if err != nil {
		return keyId, err
	}

	return responseBody.Id, err
}

func (provider GitHubProvider) RemoveSSHKey(keyId int, accessToken string) (int, error) {
	var status int
	var err error

	url := fmt.Sprintf("https://api.github.com/user/keys/%d", keyId)

	// Construct POST request
	// From: https://www.golangprograms.com/how-do-you-send-an-http-delete-request-in-go.html
	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return status, err
	}

	// Format access token header
	accessTokenHeader := fmt.Sprintf("Bearer %s", accessToken)

	// Add request headers
	req.Header.Add("Authorization", accessTokenHeader)
	req.Header.Add("Accept", "application/vnd.github+json")
	req.Header.Add("X-GitHub-Api-Version", "2022-11-28")

	// Execute add SSH key request
	res, err := client.Do(req)
	if err != nil {
		return status, err
	}
	defer res.Body.Close()

	// Read response status
	status = res.StatusCode

	return status, err
}
