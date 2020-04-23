# syncfile
自动同步数据到远程或本地的其他目录


# 环境要求
1 go version >= 1.12
2 远程的机器必须是linux 或mac
3 可以秘钥登录远程的机器



# 配置文件 config
```ini
# 本机源目录

[server]
src=D:\share
# 是否启用缓存， 调试的时候关上
load=true
# 保存的缓存文件
gob=gob.txt

[remote]
# 如果是 false， 后面的ssh 信息必须要填写， 否则无效
dst=/home/test
islocal=false
host=192.168.0.100
port=22
user=root


[log]
# 相对日志目录
path=log
size=0
# 每天备份一次日志
every=true
```

# 启动 
```
```
