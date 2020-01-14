package ssh

import (
	"fmt"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"net"
	"ngw/log"
	"os"
	"path"
)

type Ssh struct {
	IP       string
	PORT     string
	Username string
	Password string
	client   *ssh.Client
	session  *ssh.Session
	sftp     *sftp.Client
}

func (s *Ssh) RunCommand(cmd string) (rs []byte, err error) {
	se := s.NewSshConn()
	defer se.Close()
	if buf, err := se.CombinedOutput(cmd + "\n"); err != nil {
		return buf, err
	} else {
		if len(buf) == 0 {
			return []byte(fmt.Sprintf("[Run] %s successfully", cmd)), nil
		}
		return buf, nil
	}
}

func (s *Ssh) SendFile(src, dst string) error {
	l := log.GetLogger()
	sf := s.NewSftpConn()
	defer sf.Close()
	srcType, err := os.Stat(src)
	if err != nil {
		return err
	}
	if srcType.IsDir() {
		return errors.New("source path must be a file, but got a directory")
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstType, err := sf.Stat(dst)
	if err != nil {
		return err
	}
	if dstType.IsDir() {
		dst = fmt.Sprintf("%s%s", dst, path.Base(src))
	}
	l.Println(dst)
	dstFile, err := sf.OpenFile(dst, os.O_CREATE|os.O_RDWR)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	buf, err := ioutil.ReadAll(srcFile)
	if err != nil {
		return err
	}
	dstFile.Write(buf)
	return nil
}

func (s *Ssh) Close() {
	l := log.GetLogger()
	if err := s.client.Close(); err != nil {
		l.Println("close ssh client failed,", err)
		return
	}
}

func (s *Ssh) NewSshConn() *ssh.Session {
	var err error
	s.session, err = s.client.NewSession()
	if err != nil {
		return nil
	}
	return s.session
}

func (s *Ssh) NewSftpConn() *sftp.Client {
	var err error
	s.sftp, err = sftp.NewClient(s.client)
	if err != nil {
		return nil
	}
	return s.sftp
}

func NewSSH(ip, port, username, password string) (*Ssh, error) {
	l := log.GetLogger()
	rs := new(Ssh)
	keyboardInteractiveChallenge := func(
		user,
		instruction string,
		questions []string,
		echos []bool,
	) (answers []string, err error) {
		if len(questions) == 0 {
			return []string{}, nil
		}
		return []string{password}, nil
	}
	conf := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.KeyboardInteractive(keyboardInteractiveChallenge),
			ssh.Password(password),
		},
		//HostKeyCallback: ssh.FixedHostKey(hostKey),
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	client, err := ssh.Dial("tcp", net.JoinHostPort(ip, port), conf)
	if err != nil {
		l.Println(err.Error())
		panic(err.Error())
	}

	rs.IP = ip
	rs.PORT = port
	rs.Username = username
	rs.client = client
	if client == nil {
		return nil, errors.New("connecting to ssh server failed")
	}
	return rs, nil
}
