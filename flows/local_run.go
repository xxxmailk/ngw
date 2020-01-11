package flows

import (
	"fmt"
	"ngw/local"
)

// pool: 池名称
// rbd: 卷名
// size: 单位MB
// 不回滚删除rbd动作，时间太长，要回滚就在页面上直接删就好了
// 在主节点执行的命令
func CreateRbdLocal(pool, rbd string, size string) error {
	_, err := local.Run("rbd", "info", "-p", pool, rbd)
	if err != nil {
		_, err := local.Run("rbd", "create", rbd, "--size", size, "--pool", pool)
		return err
	}
	return fmt.Errorf("rbd volume %s/%s is exsited", pool, rbd)

}
