package ssh

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"

	"golang.org/x/crypto/ssh"
)

// Key generation from:
// https://gist.github.com/goliatone/e9c13e5f046e34cef6e150d06f20a34c

func GeneratePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	var privateKey *rsa.PrivateKey
	var err error

	// Generate private key from random source
	privateKey, err = rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return privateKey, err
	}

	// Validate private key
	err = privateKey.Validate()
	if err != nil {
		return privateKey, err
	}

	return privateKey, err
}

func PrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// ASN.1 DER format
	privateDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// PEM block
	privateBlock := pem.Block{
		Type: "RSA PRIVATE KEY",
		Headers: nil,
		Bytes: privateDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privateBlock)

	return privatePEM
}

func GeneratePublicKey(privateKey *rsa.PrivateKey) (*ssh.PublicKey, error) {
	// Generate public key from private key
	publicRsaKey, err := ssh.NewPublicKey(&privateKey.PublicKey)

	return &publicRsaKey, err
}

func PublicKeyToBytes(publicKey *ssh.PublicKey) []byte {
	return ssh.MarshalAuthorizedKey(*publicKey)
}

func WriteKeyToFile(keyBytes []byte, filepath string) error {
	err := os.WriteFile(filepath, keyBytes, 0600)
	return err
}
