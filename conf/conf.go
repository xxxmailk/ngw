package conf

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"ngw/log"
)

//***********配置************
var config = RefreshConfig("/etc/ngw/ngw.yml")

type Cfg struct {
	Clusters []Cluster `yaml:"clusters"`
	NfsRoot  string    `yaml:"nfs_root"`
}

type Cluster struct {
	Name  string `yaml:"name"`
	VIP   string `yaml:"vip"`
	Nodes []Node `yaml:"nodes"`
}

type Node struct {
	Name     string `yaml:"name"`
	Ip       string `yaml:"ip"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Port     string `yaml:"port"`
}

func RefreshConfig(path string) *Cfg {
	var c = new(Cfg)
	var l = log.GetLogger()
	f, err := ioutil.ReadFile(path)
	if err != nil {
		l.Fatalln("read config file failed, please check config path,", err)
	}
	err = yaml.Unmarshal(f, c)
	if err != nil {
		l.Fatalln("unmarshal config file failed, check your yaml format", err)
	}
	return c
}

func GetConfig() *Cfg {
	return config
}
