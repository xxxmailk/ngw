module ngw

go 1.13

require (
	github.com/pkg/errors v0.8.1
	github.com/pkg/sftp v1.10.1
	github.com/sirupsen/logrus v1.4.2
	golang.org/x/crypto v0.0.0-20191219195013-becbf705a915
	golang.org/x/term v0.0.0-20191110171634-ad39bd3f0407
	gopkg.in/yaml.v2 v2.2.2
)

replace github.com/xxxmailk/ngw => ../
