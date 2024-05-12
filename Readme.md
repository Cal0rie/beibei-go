# 北北Go For 鱼鹰聊天

## 请配合鱼鹰聊天一起使用
[跳转鱼鹰聊天](https://chat.haaland.top)

## 调试运行
```
go run app/main.go
```
## 编译
```
go build -o dist/beibei-go app/main.go
```
## 配置
配置文件默认为运行目录的`config.yaml`，启动中添加参数`-c 绝对路径`即可指定启动路径，如
```
./beibei-go -c ~/coder/go/beibei-go/config.yaml
```
### 配置项
1. `api-key`：必需项，智谱清言API KEY，可通过[智谱AI开放平台](https://open.bigmodel.cn/)获取
1. `max-dialogue`：最长对话轮数，默认为10
1. `chat-uuid`：鱼鹰聊天中，机器人登录所使用的uuid
#### 范例
```
api-key: xxxxxxxx.xxxx
max-dialogue: 10
chat-uuid: ef1f53e2-5259-41b2-816b-7dbc0fddace8
```