package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"slingshot/core"
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

func testSSH(hostname string, user string, password string, knownHostFiles []string){
	godotenv.Load()

	// Test node
	node := core.Node{
		Name: "Test",
		Hostname: hostname,
		User: user,
	}

	keyBitSize := 4096
	slingshotDir := ".slingshot"

	// Initialize SSH connection for node
	initialized, err := node.InitSSH(password, false, keyBitSize, knownHostFiles, slingshotDir)
	exitOnError("Unable to init SSH connection:", err)
	fmt.Println("Init: ", initialized)

	// Execute SSH command
	cmds := []string{"echo Hello!"}
	res, err := node.ExecCmds(cmds, knownHostFiles, slingshotDir)
	exitOnError("Unable to execute SSH commands:", err)
	for i, cmd := range cmds {
		fmt.Printf("$ %s > %s\n", cmd, res[i])
	}
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

	// Create HTTP client
	client := &http.Client{}

	// Add to Git provider
	keyId, err := provider.AddSSHKey(publicKeyBytes, "slingshot", accessToken, client)
	exitOnError("Unable to add SSH key to Git provider:", err)

	log.Print("Successfully added new public key to Git provider")

	fmt.Println(keyId)

	log.Print("Removing public key from Git provider...")

	// Remove key from Git provider
	success, err := provider.RemoveSSHKey(keyId, accessToken, client)
	exitOnError("Could not remove SSH key from Git provider:", err)

	log.Print(success)
}

func main() {
	godotenv.Load()

	// Test SSH
	hostname := os.Getenv("REMOTE_HOST")
	user := os.Getenv("REMOTE_USER")
	password := os.Getenv("REMOTE_PASSWORD")
	knownHostFiles := []string{os.Getenv("KNOWN_HOST_FILE")}
	testSSH(hostname, user, password, knownHostFiles)

	// Test git provider
	gitHubAccessToken := os.Getenv("GITHUB_ACCESS_TOKEN")
	testGitProviderSSH(git.GitHub, gitHubAccessToken)
}
