package ssh

import "testing"

func TestSsh_NewSshConn(t *testing.T) {
	s, err := NewSSH("192.168.56.75", "22", "root", "123.com")
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.RunCommand("cd /opt/ && touch 111")
	if err != nil {
		t.Error(err)
	}
}
