package git

import "net/http"

type GitProvider interface {
	AddSSHKey(publicKeyBytes []byte, title string, accessToken string, client *http.Client) (int, error)
	RemoveSSHKey(keyId int, accessToken string, client *http.Client) (bool, error)
}
