package git

type GitProvider interface {
	AddSSHKey(publicKeyBytes []byte, title string, accessToken string) (int, error)
	RemoveSSHKey(keyId int, accessToken string) (bool, error)
}
