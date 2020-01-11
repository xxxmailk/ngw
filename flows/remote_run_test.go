package flows

import (
	"testing"
)

func TestRepeatRbdCheck(t *testing.T) {
	str := []string{
		// %s\/%s
		"foopool/bar1    id=admin,keyring=/etc/ceph/ceph.client.admin.keyring",
		"foopool/bar2    id=admin,keyring=/etc/ceph/ceph.client.admin.keyring",
		"foopool/bar3    id=admin,keyring=/etc/ceph/ceph.client.admin.keyring",
	}
	if ck := RbdExisted(str, "bar", "foopool"); !ck {
		t.Logf("bar is not existed")
	}
	if ck := RbdExisted(str, "bar1", "foopool"); ck {
		t.Log("bar1 is existed")
	}
	if ck := RbdExisted(str, "bar1", "foopool"); ck {
		t.Log("bar1 is existed")
	}
	if ck := RbdExisted(str, "bar1", "foopool"); ck {
		t.Log("bar1 is existed")
	}
	if ck := RbdExisted(str, "wula", "foopool"); !ck {
		t.Log("wula is not existed")
	}
	t.Log(len("\n"))
}

func TestReadLineString(t *testing.T) {
	bt := `
&ok year
# hidden
#asdfdjlajfljalf
foopool/bar1    id=admin,keyring=/etc/ceph/ceph.client.admin.keyring,
foopool/bar2    id=admin,keyring=/etc/ceph/ceph.client.admin.keyring,
foopool/bar3    id=admin,keyring=/etc/ceph/ceph.client.admin.keyring,`

	t.Log(ReadLineString([]byte(bt)))

}
