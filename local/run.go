package local

import (
	"bytes"
	"ngw/log"
	"os/exec"
)

func Run(name string, args ...string) ([]byte, error) {
	l := log.GetLogger()
	l.Debugln("running command", name, args)
	var buf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stderr = &buf
	cmd.Stdout = &buf
	err := cmd.Run()
	return buf.Bytes(), err
}
