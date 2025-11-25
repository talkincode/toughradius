package sftpc

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	User      string
	Password  string
	Host      string
	Port      int
	sshClient *ssh.Client
}

func NewClient(user string, password string, host string, port int) (*Client, error) {
	c := &Client{User: user, Password: password, Host: host, Port: port}
	err := c.Connect()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Client) Connect() error {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(s.Password))

	clientConfig = &ssh.ClientConfig{
		User:            s.User,
		Auth:            auth,
		Timeout:         10 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", s.Host, s.Port)

	if s.sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return err
	}
	return nil
}

func (s *Client) Close() error {
	return s.sshClient.Close()
}

func (s *Client) ExecCommand(session *ssh.Session, cmd string) (string, error) {
	var err error
	var _session *ssh.Session
	if session == nil {
		_session, err = s.sshClient.NewSession()
		if err != nil {
			return "", err
		}
		defer _session.Close()
	} else {
		_session = session
	}
	rbyte, err := _session.CombinedOutput(cmd)
	if err != nil {
		return string(rbyte), err
	}
	return string(rbyte), err
}
