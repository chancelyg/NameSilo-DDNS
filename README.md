一个基于 golang 实现的 Namesilo 动态DNS程序

# 1. 运行
可直接运行二进制程序，也可自行编译

```bash
namesilo-ddns --domain=chancel.me --name=test --type=A --record=4.4.4.4 --key=123456
```

运行参数如下：

* **--domain**：主域名
* **--name**：子域名
* **--type**：解析类型（A/AAAA/TXT/CNAME）
* **--record**：解析的IP值
* **--key**：NameSilo的用户Key值


# 2. 二次开发
## 2.1. 环境依赖
Go版本
* go version go1.21.4 linux/amd64

环境搭建
```shell
git clone <repo>
cd <repo_name>
go get -u
```

## 2.2. Visual Studio Code开发配置
参考启动的配置文件（LAUNCH. JSON）如下
```json
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "NameSilo",
            "type": "go",
            "request": "launch",
            "mode": "auto",
            "program": "${workspaceFolder}/main.go",
            "args": [
                "--domain=chancel.me",
                "--type=A",
                "--name=test",
                "--record=4.4.4.4",
                "--key=123456",
                "--debug=true"
            ]
        }
    ]
}
```