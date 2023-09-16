package git

type GitProvider interface {
	AddSSHKey(publicKeyBytes []byte, title string, accessToken string) (int, error)
}
