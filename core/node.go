package core

import (
	"fmt"
	"os"
	"path/filepath"
	"slingshot/ssh"
)

type Node struct {
	Name string
	Hostname string
	User string
	SSHPort *int
	PrivateKeyPath *string
	PublicKeyPath *string
}

// Use user-provided or default SSH port
func (node Node) getSSHPort() int {
	if node.SSHPort != nil {
		return *node.SSHPort
	} else {
		return 22
	}
}

// Use user-provided or default path for private key
func (node Node) getPrivateKeyPath (slingshotDir string) string {
	if node.PrivateKeyPath != nil {
		return *node.PrivateKeyPath
	} else {
		return fmt.Sprintf("%s/%s/id_rsa", slingshotDir, node.Name)
	}
}

// Use user-provided or default path for public key
func (node Node) getPublicKeyPath (slingshotDir string) string {
	if node.PublicKeyPath != nil {
		return *node.PublicKeyPath
	} else {
		return fmt.Sprintf("%s/%s/id_rsa.pub", slingshotDir, node.Name)
	}
}

func (node Node) InitSSH(password string, force bool, keyBitSize int, knownHostFiles []string, slingshotDir string) (bool, error) {
	var err error

	privateKeyPath := node.getPrivateKeyPath(slingshotDir)
	publicKeyPath := node.getPublicKeyPath(slingshotDir)

	// Always perform SSH init when force is used
	doInit := force

	// Perform init if private key is missing
	_, err = os.ReadFile(privateKeyPath)
	if err != nil {
		doInit = true
	}

	// Perform init if public key is missing
	_, err = os.ReadFile(publicKeyPath)
	if err != nil {
		doInit = true
	}

	// Exit if SSH init is not performed
	if !doInit {
		return false, err
	}

	// Generate private key
	privateKey, err := ssh.GeneratePrivateKey(keyBitSize)
	if err != nil {
		return false, err
	}

	// Encode private key as PEM -> get bytes
	privateKeyBytes := ssh.PrivateKeyToPEM(privateKey)

	// Generate public key
	publicKey, err := ssh.GeneratePublicKey(privateKey)
	if err != nil {
		return false, err
	}
	publicKeyBytes := ssh.PublicKeyToBytes(publicKey)

	// Create directories (if not exist)
	err = os.MkdirAll(filepath.Dir(privateKeyPath), os.ModePerm)
	if err != nil {
		return false, err
	}
	err = os.MkdirAll(filepath.Dir(publicKeyPath), os.ModePerm)
	if err != nil {
		return false, err
	}

	// Write key pair to disk
	ssh.WriteKeyToFile(privateKeyBytes, privateKeyPath)
	ssh.WriteKeyToFile(publicKeyBytes, publicKeyPath)

	nodeSSHPort := node.getSSHPort()
	nodeHost := fmt.Sprintf("%s:%d", node.Hostname, nodeSSHPort)

	// Create SSH config for connection with password
	pwdConfig, err := ssh.GetClientPasswordConfig(nodeHost, node.User, password, knownHostFiles...)
	if err != nil {
		return false, err
	}

	// Create SSH session
	session, conn, err := ssh.GetTCPSession(nodeHost, pwdConfig)
	if err != nil {
		return false, err
	}
	defer conn.Close()
	defer session.Close()

	// Copy public key to server
	// From: https://www.educative.io/answers/how-to-add-ssh-key-to-server
	publicKeyStr := string(publicKeyBytes)
	cmd := fmt.Sprintf("echo \"%s\" >> .ssh/authorized_keys", publicKeyStr)
	_, err = ssh.ExecCmd(session, cmd)
	if err != nil {
		return false, err
	}

	return true, err
}

func (node Node) ExecCmds(cmds []string, knownHostFiles []string, slingshotDir string) ([]string, error) {
	var results []string

	// Load private key from file
	privateKeyPath := node.getPrivateKeyPath(slingshotDir)
	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return results, err
	}

	nodeSSHPort := node.getSSHPort()
	nodeHost := fmt.Sprintf("%s:%d", node.Hostname, nodeSSHPort)

	// Create SSH client config using public key authentication
	clientConf, err := ssh.GetClientPublicKeyConfig(nodeHost, node.User, privateKeyBytes, knownHostFiles...)
	if err != nil {
		return results, err
	}

	// Create SSH session
	session, conn, err := ssh.GetTCPSession(nodeHost, clientConf)
	if err != nil {
		return results, err
	}
	defer conn.Close()
	defer session.Close()

	// Execute commands and collect results
	for _, cmd := range cmds {
		res, err := ssh.ExecCmd(session, cmd)
		if err != nil {
			return results, err
		}
		results = append(results, res)
	}

	return results, err
}
