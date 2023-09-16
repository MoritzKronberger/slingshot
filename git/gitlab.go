package git

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// GitLab SSH key (POST) API (missing from official docs)
// https://stackoverflow.com/a/38164825
// https://stackoverflow.com/questions/63551637/how-to-add-ssh-key-to-gitlab-via-api

type AddSSHKeyResponseGitLab struct {
	Key string `json:"key"`
	Id int `json:"id"`
	Title string `json:"title"`
	CreatedAt string `json:"created_at"`
	ExpiresAt string `json:"expires_at"`
	UsageType string `json:"usage_type"`
}

func AddSSHKeyGitLab(publicKeyBytes []byte, title string, accessToken string) (int, error) {
	var keyId int
	var err error

	url := "https://gitlab.com/api/v4/user/keys"
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

	// Add request headers
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Private-Token", accessToken)

	// Execute add SSH key request
	res, err := client.Do(req)
	if err != nil {
		return keyId, err
	}
	defer res.Body.Close()

	// Read response body -> Get key id
	// From: https://stackoverflow.com/a/31129967
	responseBody := &AddSSHKeyResponseGitLab{}
	err = json.NewDecoder(res.Body).Decode(responseBody)
	if err != nil {
		return keyId, err
	}

	return responseBody.Id, err
}
