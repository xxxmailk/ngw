package run

import (
	"fmt"
	"ngw/conf"
	"ngw/flows"
	"ngw/log"
	"ngw/ssh"
	"regexp"
	"strings"
)

const version = "1.0"

var agentPath = "/usr/share/ngw/ngw_operator"
var sshBuf map[string]*ssh.Ssh

// 从配置中获取集群信息
// 如果没找到就直接panic
func getClusterByName(name string) conf.Cluster {
	cfg := conf.GetConfig()
	for _, v := range cfg.Clusters {
		if v.Name == name {
			return v
		}
	}
	panic("cluster name not found in configurations")
}

// 流程函数列，本地执行，远程执行及输出
// 所有函数必须满足runner handle函数要求
// runner函数必须以Run开头
// runner函数要求：func(pool, rbd string, speaker chan string) error
func createSsh(node conf.Node) *ssh.Ssh {
	s, err := ssh.NewSSH(node.Ip, node.Port, node.Username, node.Password)
	if err != nil {
		panic(fmt.Sprintf("connect to server %s failed %s", node.Name, err))
	}
	return s
}

// 初始化本次执行集群的所有ssh连接
func InitClusterSsh(cluster conf.Cluster) {
	rs := make(map[string]*ssh.Ssh)
	for _, v := range cluster.Nodes {
		rs[v.Name] = createSsh(v)
	}
	rs["VIP"] = createSsh(conf.Node{
		Name:     "VIP",
		Ip:       cluster.VIP,
		Username: cluster.Nodes[0].Username,
		Password: cluster.Nodes[0].Password,
		Port:     cluster.Nodes[0].Port,
	})
	sshBuf = rs
}

// 关闭所有ssh连接
// todo: mian函数退出时释放所有ssh
func CloseClusterSsh() {
	for _, v := range sshBuf {
		v.Close()
	}
}

// 检查agent是否存在
func checkAgentIsExist(s *ssh.Ssh) bool {
	l := log.GetLogger()
	buf, err := s.RunCommand("/usr/share/ngw/ngw_operator -version")
	if err != nil {
		return false
	}
	if string(buf) != version {
		l.Errorf("节点 %s 代理版本异常", s.IP)
		return false
	}
	l.Infof("节点代理版本为：%s", string(buf))
	return true
}

// 如果检测版本不对或命令不存在，则重新发送agent
func sendAgent(s *ssh.Ssh) error {
	l := log.GetLogger()
	l.Infof("检测节点%s代理是否正常", s.IP)
	var buf []byte
	var err error
	buf, err = s.RunCommand("if [ ! -d /usr/share/ngw ];then mkdir -p /usr/share/ngw; fi")
	if err != nil {
		l.Errorf("create directory /usr/share/ngw failed, %s", err)
		return err
	}
	l.Debugln(buf)
	if err := s.SendFile("/usr/share/ngw/ngw_operator", "/usr/share/ngw/"); err != nil {
		return err
	}
	l.Infof("正在传输agent至: %s", s.IP)
	_, err = s.RunCommand("chmod +x /usr/share/ngw/ngw_operator")
	if !checkAgentIsExist(s) {
		return fmt.Errorf("agent version check failed, send agent failed")
	}
	return err
}

// **************** 单项执行流程
// * 优先执行
// 在所有节点执行，检查代理版本
func FlowCheckAgent(pool, rbd string, args ...string) error {
	for k, v := range sshBuf {
		if k != "VIP" {
			if !checkAgentIsExist(v) {
				return sendAgent(v)
			}
		}
	}
	return nil
}

// 在本地执行，创建rbd
func FlowCreateRBD(pool, rbd string, args ...string) error {
	l := log.GetLogger()
	l.Debugf("received parameters %s", args)
	size := strings.TrimSpace(args[0])
	if args[0] == "" {
		return fmt.Errorf("rbd volume size cannot be null")
	}
	reg := regexp.MustCompile("^\\d+\\w+$")
	if !reg.MatchString(size) {
		return fmt.Errorf("rbd卷大小\"%s\"格式不正确，请输入： 1024M  10T 等", size)
	}
	l.Infoln("正在%s池中创建rbd: %s, 容量为%s, 该操作可能需要一定时间，请耐心等待", pool, rbd, size)
	return flows.CreateRbdLocal(pool, rbd, args[0])
}

// delete rbd方法不提供，删除rbd时间太长要删自己去页面上删
//func FlowDeleteRBD(pool, rbd string, speaker chan Message, args ...string) error {
//
//}

// 添加rbdmap条目 所有节点执行
func FlowAddRbdMap(pool, rbd string, args ...string) error {
	for k, v := range sshBuf {
		if k != "VIP" {
			_, err := v.RunCommand(fmt.Sprintf("%s -action AddRbdMap -pool %s -rbd %s", agentPath, pool, rbd))
			return err
		}
	}
	return nil
}

// 删除rbdmap条目 所有节点执行, add 方法的回滚
func FlowRemoveRbdMap(pool, rbd string, args ...string) error {
	for k, v := range sshBuf {
		if k != "VIP" {
			_, err := v.RunCommand(fmt.Sprintf("%s -action DeleteRbdMap -pool %s -rbd %s", agentPath, pool, rbd))
			return err
		}
	}
	return nil
}

// 在VIP节点执行
// 挂载rbd，根据参数 arg[0] == "formatConfirm"，判断是否格式化
// 确认格式化在命令中多次询问
func FlowMappingRbd(pool, rbd string, args ...string) error {
	_, err := sshBuf["VIP"].RunCommand(fmt.Sprintf("%s -action RbdMap -pool %s -rbd %s", agentPath, pool, rbd))
	if err != nil {
		return err
	}
	if args[0] == "formatConfirm" {
		_, err := sshBuf["VIP"].RunCommand(fmt.Sprintf("%s -action FormatRbd -pool %s -rbd %s", agentPath, pool, rbd))
		return err
	}
	return nil
}

// 在VIP节点执行，卸载rbd
// mappingrbd的回滚
func FlowUnMapRbd(pool, rbd string, args ...string) error {
	_, err := sshBuf["VIP"].RunCommand(fmt.Sprintf("%s -action RbdUnMap -pool %s -rbd %s", agentPath, pool, rbd))
	if err != nil {
		return err
	}
	return nil
}

// 4、添加/etc/exports 双节点 √
func FlowAddExports(pool, rbd string, args ...string) error {
	c := conf.GetConfig()
	for k, v := range sshBuf {
		if k != "VIP" {
			_, err := v.RunCommand(fmt.Sprintf(
				"%s -action AddNfsExports -nfsroot %s -pool %s -rbd %s",
				agentPath, c.NfsRoot, pool, rbd))
			return err
		}
	}
	return nil
}

// 4、删除/etc/exports 双节点 √
// 添加exports的回滚
func FlowRemoveExports(pool, rbd string, args ...string) error {
	c := conf.GetConfig()
	for k, v := range sshBuf {
		if k != "VIP" {
			_, err := v.RunCommand(fmt.Sprintf(
				"%s -action DeleteNfsExports -nfsroot %s -pool %s -rbd %s",
				agentPath, c.NfsRoot, pool, rbd))
			return err
		}
	}
	return nil
}

// 5、执行exportfs  VIP节点  √
// 根据args[0] == "force"确定是否强制重载, 如果卸载nfs的时候遇到问题，可以考虑强制重载，可能会导致客户端中断重连，理论上不影响业务
// 无回滚方法
func FlowApplyExports(pool, rbd string, args ...string) error {
	if args[0] == "force" {
		_, err := sshBuf["VIP"].RunCommand(fmt.Sprintf("%s -action ExportFsARV --force", agentPath))
		if err != nil {
			return err
		}
	} else {
		_, err := sshBuf["VIP"].RunCommand(fmt.Sprintf("%s -action ExportFsARV", agentPath))
		if err != nil {
			return err
		}
	}
	return nil
}

// 6、添加资源到pacemaker vip节点
func FlowAddResource(pool, rbd string, args ...string) error {
	_, err := sshBuf["VIP"].RunCommand(fmt.Sprintf(
		"%s -action CreateResource --pool %s --rbd %s", agentPath, pool, rbd))
	return err
}

// addResource方法的回滚
func FlowDeleteResource(pool, rbd string, args ...string) error {
	_, err := sshBuf["VIP"].RunCommand(fmt.Sprintf(
		"%s -action DeleteResource --pool %s --rbd %s", agentPath, pool, rbd))
	return err
}

// **************** 单项执行流程

func CreateVolume(clusterName, pool, rbd, size string, format, force bool) {
	l := log.GetLogger()
	l.Infoln("beginning to create rbd volume")
	r := NewRunner(pool, rbd)
	c := getClusterByName(clusterName)
	l.Debugln("got cluster info ", c)
	forceExportReload := ""
	if force {
		forceExportReload = "force"
	}
	formatConfirm := ""
	if format {
		formatConfirm = "formatConfirm"
	}
	nilFunc := func(pool, rbd string, args ...string) error { return nil }
	// 1:初始化ssh
	InitClusterSsh(c)
	// 2:检查agent
	r.Register(FlowCheckAgent, nilFunc, "准备检查节点代理", "")
	// 3:创建rbd
	r.Register(FlowCreateRBD, nilFunc, "准备创建rbd", "", size)
	// 4:添加rbdmap
	r.Register(FlowAddRbdMap, FlowRemoveRbdMap,
		"准备添加到节点RbdMap配置表",
		"[RollBack] 准备从RbdMap配置表中删除卷",
	)
	// 5:挂载rbd
	r.Register(FlowMappingRbd, FlowUnMapRbd,
		"准备挂载rbd到VIP节点",
		"[RollBack] 准备从VIP节点卸载RBD",
		formatConfirm,
	)
	// 6:添加exports
	r.Register(FlowAddExports, FlowRemoveExports,
		"准备添加卷到nfs export项目",
		"[RollBack] 准备从nfs export中删除卷",
	)
	// 7:重载nfs网关
	r.Register(FlowApplyExports, FlowApplyExports,
		"准备重载nfs网关",
		"[RollBack] 准备重载nfs网关",
		forceExportReload,
	)
	// 8:添加到pacemaker集群
	r.Register(FlowAddResource, FlowDeleteResource,
		"正在添加卷资源到HA集群",
		"[RollBack] 正在从HA集群中删除卷",
	)
	r.Run()
}

func RemoveVolume(clusterName, pool, rbd string, format, force bool) {
	r := NewRunner(pool, rbd)
	// close speaker progress
	c := getClusterByName(clusterName)
	forceExportReload := ""
	if force {
		forceExportReload = "force"
	}
	formatConfirm := ""
	if format {
		formatConfirm = "formatConfirm"
	}
	// 1:初始化ssh
	InitClusterSsh(c)
	// 2:检查agent
	r.Register(FlowCheckAgent, nil, "准备检查节点代理", "")
	// 7:从pacemaker集群删除
	r.Register(FlowDeleteResource, FlowAddResource,
		"正在从HA集群中删除卷",
		"[RollBack] 正在卷资源到HA集群",
	)
	// 5:删除exports
	r.Register(FlowRemoveExports, FlowAddExports,
		"准备添加卷到nfs export项目",
		"[RollBack] 准备从nfs export中删除卷",
	)
	// 6:重载nfs网关
	r.Register(FlowApplyExports, FlowRemoveExports,
		"准备从nfs export中删除卷",
		"[RollBack] 准备添加卷到nfs export项目",
		forceExportReload,
	)
	// 4:卸载rbd
	r.Register(FlowUnMapRbd, FlowMappingRbd,
		"准备从VIP节点卸载RBD",
		"[RollBack] 准备挂载rbd到VIP节点",
		formatConfirm,
	)
	// 3:从rbdmap删除rbd
	r.Register(FlowAddRbdMap, FlowRemoveRbdMap,
		"准备从RbdMap配置表中删除卷",
		"[RollBack] 准备添加到节点RbdMap配置表",
	)
	r.Run()
}
