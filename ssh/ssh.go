package ssh

import (
	"bytes"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// SSH session configuration from:
// https://medium.com/@marcus.murray/go-ssh-client-shell-session-c4d40daa46cd

func GetClientPasswordConfig(host string, user string, pwd string, knownHostFiles ...string) (*ssh.ClientConfig, error) {
	var clientConf *ssh.ClientConfig
	var err error

	// Create host key callback
	hostKeyCallback, err := knownhosts.New(knownHostFiles...)
	if err != nil {
		return clientConf, err
	}

	// Create client config
	clientConf = &ssh.ClientConfig{
		User: user,
		HostKeyCallback: hostKeyCallback,
		Auth: []ssh.AuthMethod{
			ssh.Password(pwd),
		},
	}

	return clientConf, err
}

func GetClientPublicKeyConfig(host string, user string, pkey []byte, knownHostFiles ...string) (*ssh.ClientConfig, error) {
	var clientConf *ssh.ClientConfig
	var err error

	// Create signer for private key
	signer, err := ssh.ParsePrivateKey(pkey)
	if err != nil {
		return clientConf, err
	}

	// Create host key callback
	hostKeyCallback, err := knownhosts.New(knownHostFiles...)
	if err != nil {
		return clientConf, err
	}

	// Create client config
	clientConf = &ssh.ClientConfig{
		User: user,
		HostKeyCallback: hostKeyCallback,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
	}

	return clientConf, err
}

func GetTCPSession(host string, clientConfig *ssh.ClientConfig) (*ssh.Session, *ssh.Client, error) {
	var session *ssh.Session
	var conn *ssh.Client
	var err error

	// Create TCP client
	conn, err = ssh.Dial("tcp", host, clientConfig)
	if err != nil {
		return session, conn, err
	}
	
	// Create new TCP session
	session, err = conn.NewSession()
	if err != nil {
		return session, conn, err
	}

	return session, conn, err
}

// SSH command execution from:
// https://stackoverflow.com/a/37680243

func ExecCmd(session *ssh.Session, cmd string) (string, error) {
	// Capture output (bytes)
	var b bytes.Buffer
    session.Stdout = &b

	// Run command
    err := session.Run(cmd)

	// Output to string
    return b.String(), err
}
