package flows

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"ngw/local"
	"ngw/log"
	"os"
	"regexp"
	"strings"
)

// 按行读取，并去掉注释和空行
func ReadLineString(data []byte) []string {
	var rs []string
	sp := bytes.Split(data, []byte("\n"))
	cmp := regexp.MustCompile("^#")
	for _, v := range sp {
		if cmp.Match([]byte(v)) {
			continue
		}
		if len(v) < 2 {
			continue
		}
		rs = append(rs, string(v))
	}
	return rs
}

// rbd: 卷名
// pool: 池名称
// 只添加rbdmap条目，不做其它操作
// 1-> 根据池和卷名检查rbd条目是否存在，如果存在不做操作
// 2-> 如果不存在，添加rbd条目，秘钥统一设置，不提供参数
func AddRbdMap(pool, rbd string) error {
	f, err := os.OpenFile("/etc/ceph/rbdmap", os.O_RDONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	str := ReadLineString(content)
	check := RbdExisted(str, pool, rbd)
	if check {
		return errors.New(fmt.Sprintf("entry is existed, %s/%s", pool, rbd))
	}
	_, err = f.WriteString(fmt.Sprintf("%s/%s    id=admin,keyring=/etc/ceph/ceph.client.admin.keyring\n", pool, rbd))
	if err != nil {
		return err
	}
	return nil
}

// 添加rbd操作回滚
func AddRbdMapRollBack(pool, rbd string) error {
	return DeleteRbdMap(pool, rbd)
}

// 删除rbdmap条目 -
func DeleteRbdMap(pool, rbd string) error {
	f, err := os.OpenFile("/etc/ceph/rbdmap", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	str := ReadLineString(content)
	//cmp := regexp.MustCompile("^" + pool + "/" + rbd + "$")
	for k, v := range str {
		var poolRbd string
		_, _ = fmt.Sscanf(v, "%s ", &poolRbd)
		if poolRbd == pool+"/"+rbd {
			// 删除元素
			str = append(str[:k], str[k+1:]...)
		}
	}
	rs := strings.Join(str, "\n")
	_, err = f.Write([]byte(rs))
	if err != nil {
		return err
	}
	return nil
}

// 根据pool检测rbd是否在rbdmap中已存在
// true： 解析失败或者值已存在
// false: 不存在
func RbdExisted(data []string, pool, rbd string) bool {
	var poolRbd, keys string
	for _, v := range data {
		_, _ = fmt.Sscanf(v, "%s %s", &poolRbd, &keys)
		if poolRbd == pool+"/"+rbd {
			return true
		}
	}
	return false
}

// 检查目录是否存在，如果不存在就创建
func dirCheck(path string) error {
	l := log.GetLogger()
	s, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.Mkdir(path, 0644)
		if err != nil {
			l.Errorf("make directory failed in function dirCheck, %s", err)
			return err
		}
	}
	if !s.IsDir() {
		err = os.Remove(path)
		if err != nil {
			l.Errorf("remove path file failed in function dirCheck, %s", err)
			return err
		}
	}
	return nil
}

// 检查确认每级目录是否存在，若不存在创建目录
func checkNfsDir(nfsRoot, pool, rbd string) error {
	if err := dirCheck(nfsRoot); err != nil {
		return err
	}
	if err := dirCheck(nfsRoot + "/" + pool); err != nil {
		return err
	}
	if err := dirCheck(nfsRoot + "/" + pool + "/" + rbd); err != nil {
		return err
	}
	return nil
}

// 添加exports条目
func AddNfsExports(nfsRoot, pool, rbd string) error {
	if err := checkNfsDir(nfsRoot, pool, rbd); err != nil {
		return err
	}
	str := fmt.Sprintf("%s/%s/%s *(rw,no_root_squash,sync,no_subtree_check)\n", nfsRoot, pool, rbd)
	f, err := os.OpenFile("/etc/exports", os.O_RDWR|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(str)
	return err
}

func DeleteNfsExports(nfsRoot, pool, rbd string) error {
	f, err := os.OpenFile("/etc/exports", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	content, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	str := ReadLineString(content)
	//cmp := regexp.MustCompile("^" + pool + "/" + rbd)
	for k, v := range str {
		var poolRbd string
		_, _ = fmt.Sscanf(v, "%s ", &poolRbd)
		if poolRbd == fmt.Sprintf("%s/%s/%s", nfsRoot, pool, rbd) {
			// 删除元素
			str = append(str[:k], str[k+1:]...)
		}
	}
	rs := strings.Join(str, "\n")
	_, err = f.Write([]byte(rs))
	if err != nil {
		return err
	}
	return nil
}

// 命令挂载rbd
func RbdMap(pool, rbd string) error {
	cmd := fmt.Sprintf("rbd map %s/%s"+
		" --id admin --keyring /etc/ceph/ceph.client.admin.keyring", pool, rbd)
	_, err := local.Run(cmd)
	return err
}

// 命令卸载rbd
func RbdUnMap(pool, rbd string) error {
	cmd := fmt.Sprintf("rbd unmap %s/%s", pool, rbd)
	_, err := local.Run(cmd)
	return err
}

// 格式化rbd
// 执行该函数时一定要注意！会格式化数据，慎重！
func FormatRbd(pool, rbd string) error {
	path := fmt.Sprintf("/dev/rbd/%s/%s", pool, rbd)
	_, err := os.Stat(path)
	if err != nil {
		// 检查rbd设备文件是否存在。如果不存在，重新尝试进行map, 如果map报错，直接返回报错
		if os.IsNotExist(err) {
			if err := RbdMap(pool, rbd); err != nil {
				return err
			}
		}
		return err
	}
	_, err = local.Run("mkfs.xfs " + path)
	return err
}

// 执行exportfs -ar 强制重载nfs导出项
// 如果force==true，哪怕有激活的客户端都会强制重载
func ExportFsARV(force bool) error {
	var err error
	if force {
		_, err = local.Run("exportfs", "-arfv")
	} else {
		_, err = local.Run("exportfs", "-arv")
	}
	return err
}
