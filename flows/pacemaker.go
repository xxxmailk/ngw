package flows

import (
	"fmt"
	"github.com/pkg/errors"
	"ngw/local"
	"strings"
)

// group name: grp_nfs_rbd
const groupName = "grp_nfs_rbd"

// primitive name: res_$pool_$rbd
// mount path: /ngw/
// primitive res_qz_QZshipin1 Filesystem \
// 		params device="/dev/rbd/QZshipin_meta/QZshipin1" directory="/ngw/QZshipin1" fstype="xfs"

// 删除资源
func deletePrimitive(name string) error {
	_, err := local.Run("crm", "configure", "delete", name)
	return err
}

// 创建资源
func createPrimitive(name string, items ...string) error {
	_, err := local.Run("crm configure primitive "+name, items...)
	return err
}

func checkVolumeIsExsit(pool, rbd string) bool {
	if _, err := local.Run("crm", "configure", "show", fmt.Sprintf("res_%s_%s", pool, rbd)); err != nil {
		return false
	}
	return true
}

func verify() error {
	_, err := local.Run("crm", "configure", "verify")
	return err
}

func commit() error {
	_, err := local.Run("crm", "configure", "commit")
	return err
}

// 创建卷映射
func CreateResource(pool, rbd string) error {
	const groupName = "grp_nfs"
	//1 -> 检查选项是否存在
	//2 -> 存在则删除再重建
	//3 -> 不存在直接新建
	//4 -> 添加资源名称到组列表中段
	//5 -> 提交变更
	// 查看资源 如果不存在，直接创建
	if checkVolumeIsExsit(pool, rbd) {
		err := DeleteResource(pool, rbd)
		if err != nil {
			return err
		}
		err = createPrimitive(fmt.Sprintf("res_%s_%s", pool, rbd),
			"Filesystem",
			"params",
			"device="+fmt.Sprintf("\"/dev/rbd/%s/%s\"", pool, rbd),
			"directory="+fmt.Sprintf("\"/ngw/%s/%s\"", pool, rbd),
			"fstype=\"xfs\"",
		)
		return err
	}
	err := createPrimitive(fmt.Sprintf("res_%s_%s", pool, rbd),
		"Filesystem",
		"params",
		"device="+fmt.Sprintf("\"/dev/rbd/%s/%s\"", pool, rbd),
		"directory="+fmt.Sprintf("\"/ngw/%s/%s\"", pool, rbd),
		"fstype=\"xfs\"")
	if err != nil {
		return err
	}
	// 添加rbd到组
	if err := addResourceBeforeNfsServerToGroup(pool, rbd); err != nil {
		return err
	}
	if err := verify(); err != nil {
		return err
	}
	// 添加资源名称到组i中间
	return commit()
}

// 删除nfs卷映射
func DeleteResource(pool, rbd string) error {
	// 1 -> 从group里删除组
	// 2 -> 删除group
	// 3 -> 创建新的group
	// 4 -> 删除资源
	// 5 -> commit
	if !checkVolumeIsExsit(pool, rbd) {
		return errors.New(fmt.Sprintf("resource res_%s_%s is not exsited", pool, rbd))
	}
	// 从组内删除资源
	if err := removeResourceFromGroup(pool, rbd); err != nil {
		return err
	}
	// 停止资源
	if err := stopResource(pool, rbd); err != nil {
		return err
	}
	if err := deletePrimitive(fmt.Sprintf("res_%s_%s", pool, rbd)); err != nil {
		return err
	}
	if err := verify(); err != nil {
		return err
	}
	return commit()
}

// 获取组最后一个资源
func getLastResourceFromGroup() (string, error) {
	out, err := local.Run("crm", "configure", "show", groupName)
	if err != nil {
		return "", err
	}
	// 只取第一行
	line := strings.Split(string(out), "\n")[0]
	// 去除末尾最后一个反斜杠
	if line[len(line)] == '\\' {
		line = line[:len(line)-1]
	}
	list := strings.Split(line, " ")
	return list[len(list)-1], err
}

//  如果返回值为负数，表明该slice中不存在该该值
func SliceSearchString(s []string, key string) (rs int) {
	var low, high int
	low = 0
	high = len(s) - 1
	for low <= high {
		if s[low] == key {
			return low
		} else {
			low += 1
		}
		if s[high] == key {
			return high
		} else {
			high -= 1
		}
	}
	return -1
}

// 添加资源名称到组的倒数第二位
// 倒数第一位是res_system_nfs_server
func addResourceBeforeNfsServerToGroup(pool, rbd string) error {
	lastRes, err := getLastResourceFromGroup()
	if err != nil {
		return err
	}
	_, err = local.Run("crm", "configure",
		"modgroup", groupName,
		"add", fmt.Sprintf("res_%s_%s", pool, rbd),
		"before", lastRes)
	return err
}

// 从组内删除资源
func removeResourceFromGroup(pool, rbd string) error {
	_, err := local.Run("crm", "configure", "modgroup", groupName,
		"remove", fmt.Sprintf("res_%s_%s", pool, rbd))
	return err
}

// 停止资源
func stopResource(pool, rbd string) error {
	_, err := local.Run("crm", "resource", "stop", fmt.Sprintf("res_%s_%s", pool, rbd))
	return err
}
