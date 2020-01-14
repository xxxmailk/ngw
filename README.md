# ngw 
> ceph+rbd+nfs+pacemaker 自动创建nfs挂载的小工具

## 使用方法
#### + [在release中下载包](https://github.com/xxxmailk/ngw/releases)
#### + 上传至服务器某处并解压
#### + 进入ngw目录
#### + 安装
```shell
make install

```
#### + 修改/etc/ngw/ngw.yml配置集群列表
```yaml
# nfs根目录位置
nfs_root: "/ngw"
clusters:
  # 集群名称
  - name: cluster1
    # VIP 地址
    vip: 192.168.1.11
    # node 列表
    nodes:
      # 节点名称
      - name: node12
        # 节点IP
        ip: 192.168.1.12
        # 节点ssh用户名(尽量配置root)
        username: root
        # 节点ssh密码
        password: xxxx
        # 节点ssh端口
        port: 22

      - name: node13
        ip: 192.168.1.13
        password: xxxx
        port: 22

  - name: cluster2
    vip: 192.168.1.11
    nodes:
      - name: node12
        ip: 192.168.1.12
        username: root
        password: xxxx
        port: 22

      - name: node13
        ip: 192.168.1.13
        password: xxxx
        port: 22 
```
#### + 创建卷
```shell
ngw --create --cluster cluster1 --pool test-pool --rbd rbdname
```
#### + 删除卷
```shell
ngw --delete --cluster cluster1 --pool test-pool --rbd rbdname
```

> 
