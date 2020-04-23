# syncfile
自动同步数据到远程或本地的其他目录


# 环境要求

1 go version >= 1.12  
2 远程的机器必须是linux 或mac  
3 可以秘钥登录远程的机器  

# 注意点

-  因为文件名有空格和反斜杠的会导致路径问题， 已经默认删除了了反斜杠和空格  
-  附带了缓存， 线上使用请开启，拷贝后的文件不会重复传输 断线会保存到文件
-  没上传完成的文件不会传输
-  判断源文件是否上传完成的依据是 时间， 如果中途卡掉了， 脚本会认为已经是完整的文件， 会拷贝至同步的地方， 删除文件重新上传就可以了

# 配置文件 config
```ini
# 本机源目录

[server]
src=D:\share
# 是否启用缓存， 调试的时候关上
load=true
# 保存的缓存文件
gob=gob.txt
# 只拷贝文件名包含这个字符的文件， 空则拷贝所有
include=[]

[remote]
# 如果是 false， 后面的ssh 信息必须要填写， 否则无效
dst=/home/test
islocal=false
host=192.168.0.100
port=22
user=root
# 所属用户和用户组， 只有远程服务才生效
owner=root

[log]
# 相对日志目录
path=log
size=0
# 每天备份一次日志
every=true
```

# 启动 
```
go run main.go
```

# 打包二进制
给不同系统打包
```
export GOOS=windows  (linux 打包windows
$env:GOOS="linux"  (windows 打包 linux)  
```
```

go build main.go
```
