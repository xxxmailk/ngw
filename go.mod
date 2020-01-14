module ngw

go 1.13

require (
	github.com/pkg/errors v0.8.1
	github.com/pkg/sftp v1.10.1
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/crypto v0.0.0-20191219195013-becbf705a915
	golang.org/x/sys v0.0.0-20191026070338-33540a1f6037 // indirect
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/xxxmailk/ngw => ../
