package ssh

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

type Client struct {
	sshClient    *ssh.Client
	forwardAgent bool
}

func NewAgentClient(user, host string) (*Client, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK is empty; there seems to be no agent running")
	}
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	agentClient := agent.NewClient(conn)
	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeysCallback(agentClient.Signers),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, err
	}

	return &Client{
		sshClient: client,
	}, nil

}

func NewClient(user, host string, forwardAgent bool) (*Client, error) {
	socket := os.Getenv("SSH_AUTH_SOCK")
	if socket == "" {
		return nil, fmt.Errorf("SSH_AUTH_SOCK is empty; there seems to be no agent running")
	}
	conn, err := net.Dial("unix", socket)
	if err != nil {
		return nil, err
	}
	agentClient := agent.NewClient(conn)
	signers, err := agentClient.Signers()
	if err != nil {
		return nil, err
	}

	// key, err := ioutil.ReadFile(privKey)
	// if err != nil {
	// 	log.Fatalf("Could not get private key: %s", err)
	// }
	// var signer ssh.Signer
	// if privKeyPassphrase == "" {
	// 	signer, err = ssh.ParsePrivateKey(key)
	// } else {
	// 	signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(privKeyPassphrase))
	// }
	// if err != nil {
	// 	log.Fatalf("Could not parse private key: %s", err)
	// }

	config := &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			// ssh.PublicKeys(signer),
			ssh.PublicKeys(signers...),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	client, err := ssh.Dial("tcp", host, config)
	if err != nil {
		return nil, err
	}

	if forwardAgent {
		if err := agent.ForwardToAgent(client, agentClient); err != nil {
			return nil, err
		}
	}

	return &Client{
		sshClient:    client,
		forwardAgent: forwardAgent,
	}, nil
}

func (c *Client) Close() error {
	return c.sshClient.Close()
}

func (c *Client) TransferFile(srcName, destName string) {
	data, err := ioutil.ReadFile(srcName)
	if err != nil {
		log.Fatalf("Error reading file %s: %s", srcName, err)
	}

	sess, err := c.sshClient.NewSession()
	if err != nil {
		log.Fatalf("Could not create session: %s", err)
	}
	defer sess.Close()
	if c.forwardAgent {
		agent.RequestAgentForwarding(sess)
	}

	go func() {
		stdout, err := sess.StdoutPipe()
		if err != nil {
			log.Fatalf("Could not get stdout: %s", err)
		}
		io.Copy(os.Stdout, stdout)
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		w, err := sess.StdinPipe()
		if err != nil {
			log.Fatalf("Error getting stdin pipe: %s", err)
		}
		defer w.Close()
		if n, err := fmt.Fprintln(w, "C0700", len(data), destName); err != nil {
			log.Fatalf("Error writing: %s. Wrote %d bytes", err, n)
		}
		if n, err := w.Write(data); err != nil {
			log.Fatalf("Error writing: %s. Wrote %d bytes", err, n)
		}

		if _, err := fmt.Fprintf(w, "\x00"); err != nil {
			log.Fatalf("Error sending termination: %s", err)
		}
	}()

	go func() {
		defer wg.Done()
		err := sess.Run("/usr/bin/scp -t .")
		if err != nil {
			log.Fatalf("Error running scp: %s", err)
		}
	}()

	wg.Wait()
}

func (c *Client) Execute(cmd string) (io.Reader, io.Reader, error) {
	sess, err := c.sshClient.NewSession()
	if err != nil {
		return nil, nil, err
	}
	defer sess.Close()
	if c.forwardAgent {
		agent.RequestAgentForwarding(sess)
	}

	stdout, err := sess.StdoutPipe()
	if err != nil {
		return nil, nil, err
	}
	stderr, err := sess.StderrPipe()
	if err != nil {
		return nil, nil, err
	}

	if err = sess.Run(cmd); err != nil {
		return stdout, stderr, fmt.Errorf("Error executing %s: %s", cmd, err)
	}

	return stdout, stderr, nil
}
