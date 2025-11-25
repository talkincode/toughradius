package sftpc

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	User     string
	Password string
	Host     string
	Port     int
}

// SftpSession represents an SFTP session
type SftpSession struct {
	Client *sftp.Client
}

// NewSession creates a new SFTP session
func (s *Client) NewSession() (*SftpSession, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
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

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return &SftpSession{
		Client: sftpClient,
	}, nil
}

// Upload transfers a local file to the remote host
func (s *Client) Upload(session *SftpSession, localFilePath, remotePath string) error {
	var err error
	var _session *SftpSession
	if session == nil {
		_session, err = s.NewSession()
		if err != nil {
			return err
		}
		defer func() { _ = _session.Client.Close() }() //nolint:errcheck
	} else {
		_session = session
	}
	srcFile, err := os.Open(localFilePath)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }() //nolint:errcheck

	dstFile, err := _session.Client.Create(remotePath)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }() //nolint:errcheck

	buf := make([]byte, 1024)
	for {
		n, e := srcFile.Read(buf)
		if e != nil && e != io.EOF {
			return e
		}
		if n == 0 {
			break
		}
		if _, e := dstFile.Write(buf[:n]); e != nil {
			return e
		}
	}

	return nil
}

// Download retrieves a remote file to the local host
func (s *Client) Download(session *SftpSession, remotePath, localFilePath string) error {
	var err error
	var _session *SftpSession
	if session == nil {
		_session, err = s.NewSession()
		if err != nil {
			return err
		}
		defer func() { _ = _session.Client.Close() }() //nolint:errcheck
	} else {
		_session = session
	}

	srcFile, err := _session.Client.Open(remotePath)
	if err != nil {
		return err
	}
	defer func() { _ = srcFile.Close() }() //nolint:errcheck

	dstFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer func() { _ = dstFile.Close() }() //nolint:errcheck

	if _, err = srcFile.WriteTo(dstFile); err != nil {
		return err
	}

	return nil
}

func (s *Client) Remove(session *SftpSession, remotePath string) error {
	var err error
	var _session *SftpSession
	if session == nil {
		_session, err = s.NewSession()
		if err != nil {
			return err
		}
		defer func() { _ = _session.Client.Close() }() //nolint:errcheck
	} else {
		_session = session
	}

	err = _session.Client.Remove(remotePath)
	if err != nil {
		return err
	}
	return nil
}

func (s *Client) Close(session *SftpSession) error {
	return session.Client.Close()
}
