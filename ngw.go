package main

import (
	"flag"
	"fmt"
	"ngw/conf"
	"ngw/run"
	"os"
)

//**********创建************
// 1、创建rbd（已有池） √
// 2、修改rbdmap 双节点 √
// 3、挂载rbd（需要确认是否格dir化） vip节点 mkfs.xfs /dev/rbd/poolname/rbdname (两次确认) √
// 4、修改/etc/exports 双节点 √
// 5、执行exportfs  vip节点  √
// 6、修改pacemaker vip节点

//***********停用卷***********
//
// 1、删除/etc/exports条目 双节点
// 2、执行exportfs使其生效 vip节点
// 3、删除rbd条目 双节点
// 4、卸载rbd vip节点
// 5、删除pacemaker条目 vip节点

//***********配置************
// format: yaml
// cluster：
//   vip: 192.168.1.11
//   nodes:
//     - node12:
//         ip: 192.168.1.12
//         username: root
//         password: xxxx
//         port: 22
//     - node13:
//         ip: 192.168.1.13
//         username: root
//         password: xxxx
//         port: 22

//***********exports********
// /mnt/ec1 *(rw,no_root_squash,sync,no_subtree_check)
// /mnt/ec6/ec1 *(rw,no_root_squash,sync,no_subtree_check)
// /mnt/db/d1 *(rw,no_root_squash,sync,no_subtree_check)

func main() {
	var (
		f_cluster = flag.String("cluster", "", "要添加卷到哪个集群")
		f_tree    = flag.Bool("tree", false, "列出所有集群")
		f_create  = flag.Bool("create", false, "创建nfs卷时使用该参数")
		f_delete  = flag.Bool("delete", false, "删除nfs卷时使用该参数")
		f_pool    = flag.String("pool", "", "添加卷到哪个池")
		f_rbd     = flag.String("rbd", "", "所要添加卷的名称")
		f_size    = flag.String("size", "", "需要添加卷的大小，例如： 1G, 1000T")
		f_format  = flag.Bool("format", false, "如果是首次创建，请加上该参数，该参数会格式化rbd卷，如果有数据请勿使用该参数")
		f_force   = flag.Bool("force", false, "是否强制重载nfs网关，如果遇到nfs网关因为有活跃客户端重载失败时，可尝试强制重载")
	)
	//func CreateVolume(clusterName, pool, rbd, size string, format, force bool)
	flag.Parse()
	if *f_create {
		if *f_cluster == "" {
			fmt.Println("error: cluster name cannot be null")
			os.Exit(2)
		}
		if *f_pool == "" {
			fmt.Println("error: pool name cannot be null")
			os.Exit(2)
		}
		if *f_rbd == "" {
			fmt.Println("error: rbd name cannot be null")
			os.Exit(2)
		}
		run.CreateVolume(*f_cluster, *f_pool, *f_rbd, *f_size, *f_format, *f_force)
	} else if *f_delete {
		if *f_cluster == "" {
			fmt.Println("error: cluster name cannot be null")
			os.Exit(2)
		}
		if *f_pool == "" {
			fmt.Println("error: pool name cannot be null")
			os.Exit(2)
		}
		if *f_rbd == "" {
			fmt.Println("error: rbd name cannot be null")
			os.Exit(2)
		}
		run.RemoveVolume(*f_cluster, *f_pool, *f_rbd, *f_format, *f_force)
	} else if *f_tree {
		c := conf.GetConfig()
		fmt.Println("Tree node of nfs gateway:")
		for _, v := range c.Clusters {
			fmt.Printf("-> %s:\n", v.Name)
			for _, v1 := range v.Nodes {
				fmt.Printf("%15s\n", v1.Name)
			}
		}
	} else {
		flag.Usage()
	}
}
