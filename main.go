package main

import (
	"fmt"
	"log"
	"os"
	"slingshot/git"
	"slingshot/ssh"

	"github.com/joho/godotenv"
)

func exitOnError(msg string, err error) {
	if err != nil {
		log.Fatal(msg + " ", err)
		os.Exit(1)
	}
}

func testSHH(){
	// SSH config
	godotenv.Load()
	host := os.Getenv("REMOTE_HOST")
	user := os.Getenv("REMOTE_USER")
	password := os.Getenv("REMOTE_PASSWORD")
	knownHostFiles := []string{os.Getenv("KNOWN_HOST_FILE")}

	// Generate public and private key
	privateKeyPath := "./id_rsa"
	publicKeyPath := "./id_rsa.pub"
	keyBitSize := 4096

	privateKeyBytes, err := os.ReadFile(privateKeyPath)

	if err != nil {
		log.Print("Could not find private key, generating new key pair...")

		// Generate private key
		privateKey, err := ssh.GeneratePrivateKey(keyBitSize)
		exitOnError("Unable to generate private key:", err)

		// Encode private key as PEM -> get bytes
		privateKeyBytes = ssh.PrivateKeyToPEM(privateKey)

		// Generate public key
		publicKey, err := ssh.GeneratePublicKey(privateKey)
		exitOnError("Unable to generate public key:", err)
		publicKeyBytes := ssh.PublicKeyToBytes(publicKey)

		// Write key pair to disk
		ssh.WriteKeyToFile(privateKeyBytes, privateKeyPath)
		ssh.WriteKeyToFile(publicKeyBytes, publicKeyPath)

		log.Print("Successfully generated new key pair")
		log.Print("Copying new public key to server...")

		// Create SSH config for connection with password
		pwdConfig, err := ssh.GetClientPasswordConfig(host, user, password, knownHostFiles...)
		exitOnError("Unable to create SSH password config:", err)

		// Create SSH session
		session, conn, err := ssh.GetTCPSession(host, pwdConfig)
		exitOnError("Unable to create SHH password session:", err)
		defer conn.Close()
		defer session.Close()

		// Copy public key to server
		// From: https://www.educative.io/answers/how-to-add-ssh-key-to-server
		publicKeyStr := string(publicKeyBytes)
		cmd := fmt.Sprintf("echo \"%s\" >> .ssh/authorized_keys", publicKeyStr)
		_, err = ssh.ExecCmd(session, cmd)
		exitOnError("Unable to copy public key to server:", err)

		log.Print("Successfully copied new public key to server")
	}

	// Create SSH client config using public key authentication
	clientConf, err := ssh.GetClientPublicKeyConfig(host, user, privateKeyBytes, knownHostFiles...)
	exitOnError("Unable to create public key client config:", err)

	// Create SSH session
	session, conn, err := ssh.GetTCPSession(host, clientConf)
	exitOnError("Unable to create public key SSH session:", err)
	defer conn.Close()
	defer session.Close()

	// Execute command
	res, err := ssh.ExecCmd(session, "echo Hello!")
	exitOnError("Unable to execute command:", err)

	fmt.Println(res)
}

func testGitProviderSSH(provider git.GitProvider, accessToken string) {
	// Generate new SSH key pair for Git provider
	privateKeyPath := "./id_rsa_git"
	publicKeyPath := "./id_rsa_git.pub"
	keyBitSize := 4096

	// Generate private key
	privateKey, err := ssh.GeneratePrivateKey(keyBitSize)
	exitOnError("Unable to generate private key:", err)

	// Encode private key as PEM -> get bytes
	privateKeyBytes := ssh.PrivateKeyToPEM(privateKey)

	// Generate public key
	publicKey, err := ssh.GeneratePublicKey(privateKey)
	exitOnError("Unable to generate public key:", err)
	publicKeyBytes := ssh.PublicKeyToBytes(publicKey)

	// Write key pair to disk
	ssh.WriteKeyToFile(privateKeyBytes, privateKeyPath)
	ssh.WriteKeyToFile(publicKeyBytes, publicKeyPath)

	log.Print("Successfully generated new key pair")
	log.Print("Adding new public key to Git provider...")

	// Add to Git provider
	keyId, err := provider.AddSSHKey(publicKeyBytes, "slingshot", accessToken)
	exitOnError("Unable to add SSH key to Git provider:", err)

	log.Print("Successfully added new public key to Git provider")

	fmt.Println(keyId)

	log.Print("Removing public key from Git provider...")

	// Remove key from Git provider
	status, err := provider.RemoveSSHKey(keyId, accessToken)
	exitOnError("Could not remove SSH key from Git provider:", err)

	log.Print(status)
}

func main() {
	// Git provider config
	godotenv.Load()
	gitHubAccessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	// gitLabAccessToken := os.Getenv("GITLAB_ACCESS_TOKEN")

	testGitProviderSSH(git.GitHub, gitHubAccessToken)
}
