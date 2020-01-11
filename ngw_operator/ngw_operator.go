package main

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"ngw/flows"
	"os"
)

const version = "1.0"

func main() {
	const (
		addRbdMapErr = iota + 128
		addRbdMapRollBackErr
		deleteRbdMapErr
		addNfsExportsErr
		deleteNfsExportsErr
		rbdMapErr
		formatRbdErr
		exportFsARVErr
		createResourceErr
		deleteResourceErr
		rbdUnmapErr
	)
	var (
		actDescribe = `可执行的动作：
AddRbdMap		添加rbdmap条目 --pool --rbd
DeleteRbdMap	 	删除rbd条目 --pool --rbd
AddNfsExports	 	添加nfs条目 --nfsroot --pool --rbd
DeleteNfsExports	删除nfs条目 --nfsroot --pool --rbd
ExportFsARV		使nfs配置生效 (--force 强制重载所有nfs条目)
RbdMap			手动挂载rbd --pool --rbd
RbdUnMap			手动卸载rbd --pool --rbd
FormatRbd		格式化rbd --pool --rbd
CreateResource		创建pacemaker资源,会自动添加到组 --pool --rbd
DeleteResource  	删除pacemaker资源,同上 --pool --rbd
`
		// action
		action = flag.String("action", "Version", actDescribe)
		// pool name
		_pool = flag.String("pool", "", "需要创建的池名称")
		// rbd name
		_rbd = flag.String("rbd", "", "需要创建的rbd名称")
		// _nfsroot
		_nfsRoot = flag.String("nfsroot", "/ngw", "指定nfs根目录")
		// _force
		_force = flag.Bool("force", false, "是否强制重载nfs")
		// _help
		_help = flag.Bool("help", false, "显示帮助")
		// _version
		_version = flag.Bool("version", false, "显示执行端版本")
	)
	flag.Parse()
	if *_version{
		fmt.Print(version)
		os.Exit(0)
	}
	if *_help {
		fmt.Println("command at least have 2 arguments")
		flag.Usage()
		os.Exit(0)
	}
	switch *action {
	case "Version":
		fmt.Print(version)
		break
	case "AddRbdMap":
		if rbdPoolIsNotNil(*_pool, *_rbd) == nil{
			err := flows.AddRbdMap(*_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(addRbdMapErr)
			}
		}
		break
	case "AddRbdMapRollBack":
		if rbdPoolIsNotNil(*_pool, *_rbd) == nil{
			err := flows.AddRbdMapRollBack(*_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(addRbdMapRollBackErr)
			}
		}
		break
	case "DeleteRbdMap":
		if rbdPoolIsNotNil(*_pool, *_rbd) ==nil{
			err := flows.DeleteRbdMap(*_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(deleteRbdMapErr)
			}
		}
		break
	case "AddNfsExports":
		if rbdPoolIsNotNil(*_pool, *_rbd) == nil{
			err := flows.AddNfsExports(*_nfsRoot, *_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(addNfsExportsErr)
			}
		}
		break
	case "DeleteNfsExports":
		if rbdPoolIsNotNil(*_pool, *_rbd) == nil {
			err := flows.DeleteNfsExports(*_nfsRoot, *_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(deleteNfsExportsErr)
			}
		}
		break
	case "RbdMap":
		if rbdPoolIsNotNil(*_pool, *_rbd) == nil {
			err := flows.RbdMap(*_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(rbdMapErr)
			}
		}
		break
	case "RbdUnMap":
		if rbdPoolIsNotNil(*_pool, *_rbd) == nil {
			err := flows.RbdUnMap(*_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(rbdUnmapErr)
			}
		}
		break
	case "FormatRbd":
		if rbdPoolIsNotNil(*_pool, *_rbd) == nil {
			err := flows.FormatRbd(*_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(formatRbdErr)
			}
		}
		break
	case "ExportFsARV":
			err := flows.ExportFsARV(*_force)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(exportFsARVErr)
			}
		break
	case "CreateResource":
		if rbdPoolIsNotNil(*_pool, *_rbd) == nil{
			err := flows.CreateResource(*_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(createResourceErr)
			}
		}
		break
	case "DeleteResource":
		if rbdPoolIsNotNil(*_pool, *_rbd)==nil {
			err := flows.DeleteResource(*_pool, *_rbd)
			if err != nil {
				_, _ = io.WriteString(os.Stderr, err.Error())
				os.Exit(deleteResourceErr)
			}
		}
		break
	default:
		fmt.Println("[error] invalid action")
		flag.Usage()
	}
	os.Exit(0)
}

func rbdPoolIsNotNil(pool, rbd string) error{
	if pool == "" || rbd == ""{
		_,_=io.WriteString(os.Stderr, "argument value of pool or rbd cannot be nil")
		return errors.New("argument value of pool or rbd cannot be nil")
	}
	return nil
}
