package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"beibei/app/service"

	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

type websocketClientManager struct {
	conn        *websocket.Conn
	addr        *string
	path        string
	sendMsgChan chan string
	recvMsgChan chan string
	isAlive     bool
	timeout     int
}

type NewRequest struct {
	Model    string               `json:"model"`
	Messages []service.ReqMessage `json:"messages"`
}

// http响应
type Response struct {
	Result  string   `json:"result"`
	Message []string `json:"message"`
}

// websocket请求
type SocketRequest struct {
	Code int    `json:"code"`
	Uuid string `json:"uuid"`
	Msg  string `json:"msg"`
}

// webSocket响应
type SocketResponse struct {
	Msg      string `json:"msg"`
	UserName string `json:"userName"`
}

// 全局变量，存储多个对话
var reqData NewRequest
var configPath string

// 构造函数
func NewWsClientManager(addrIp, addrPort, path string, timeout int) *websocketClientManager {
	addrString := addrIp + ":" + addrPort
	var sendChan = make(chan string, 10) //定义channel大小，需要及时处理消费，否则会阻塞
	var recvChan = make(chan string, 10) //定义channel大小，需要及时处理消费，否则会阻塞
	var conn *websocket.Conn
	return &websocketClientManager{
		addr:        &addrString,
		path:        path,
		conn:        conn,
		sendMsgChan: sendChan,
		recvMsgChan: recvChan,
		isAlive:     false,
		timeout:     timeout,
	}
}

// 链接服务端
func (wsc *websocketClientManager) dail() {
	var err error
	u := url.URL{Scheme: "ws", Host: *wsc.addr, Path: wsc.path}
	fmt.Println("connecting to:", u.String())
	wsc.conn, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	wsc.isAlive = true
	log.Printf("connecting to %s 链接成功！！！", u.String())
	wsc.sendMsgThread("你的林北北已连接")
}

// 发送消息到服务端
func (wsc *websocketClientManager) sendMsgThread(m string) {
	// m := <-wsc.sendMsgChan
	fmt.Println("发送消息:", m)

	socketRequest := SocketRequest{
		Code: 1,
		Uuid: viper.GetString("chat-uuid"),
		Msg:  m,
	}
	// websocket.TextMessage类型
	jsonData, err := json.Marshal(socketRequest)
	// fmt.Println(string(jsonData))
	if err != nil {
		fmt.Println(err)
	}
	err = wsc.conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		fmt.Println("write:", err)
		// continue
	}
}

// 读取服务端消息
func (wsc *websocketClientManager) readMsgThread() {
	go func() {
		for {
			if wsc.conn != nil {
				_, message, err := wsc.conn.ReadMessage()
				if err != nil {
					log.Println("readErr:", err)
					wsc.isAlive = false
					// 出现错误，退出读取，尝试重连
					break
				}
				// 需要读取数据，不然会阻塞
				wsc.recvMsgChan <- string(message)

			}
		}
	}()
}

// 开启服务并重连
func (wsc *websocketClientManager) start() {
	for {
		if wsc.isAlive == false {
			wsc.dail()
			// wsc.sendMsgThread()
			wsc.readMsgThread()
			wsc.Msg()  //构造假消息
			wsc.Recv() //接收处理服务端返回到消息
		}
		time.Sleep(time.Second * time.Duration(wsc.timeout))
	}
}

// 模拟websocket心跳包，假数据
func (wsc *websocketClientManager) Msg() {
	go func() {
		a := 0
		for {
			wsc.sendMsgChan <- strconv.Itoa(a)
			time.Sleep(time.Second * 1)
			a += 1
		}
	}()
}

// 接收处理服务端返回到消息
func (wsc *websocketClientManager) Recv() {
	go func() {
		for {
			msg, ok := <-wsc.recvMsgChan
			if ok {
				fmt.Println("收到消息：", msg)
				// 检测消息中是否包含"@林北北"
				if strings.Contains(msg, "@林北北") {
					// 发送http请求，调用机器人接口
					var socketResponse SocketResponse
					err := json.Unmarshal([]byte(msg), &socketResponse)
					if err != nil {
						fmt.Println(err)
					}
					outputString := strings.ReplaceAll(socketResponse.Msg, "@林北北", "")
					wsc.Post(outputString, socketResponse.UserName)
				}
			}
		}
	}()
}

// 发送POST请求
func (wsc *websocketClientManager) Post(msg string, usr string) {
	go func() {
		targetUrl := "https://open.bigmodel.cn/api/paas/v4/chat/completions"
		fmt.Println(msg)
		// reqData := NewRequest{
		// 	Session_id: "friend-123",
		// 	Username:   usr,
		// 	Message:    msg,
		// }

		reqData.Messages = service.ShiftTheMessages(reqData.Messages)
		// 向请求参数的数组中添加消息
		reqData.Messages = append(reqData.Messages, service.ReqMessage{
			Role:    "user",
			Content: msg,
		})

		jsonData, err := json.Marshal(reqData)
		if err != nil {
			fmt.Println(err)
		}

		payload := strings.NewReader(string(jsonData))

		req, err := http.NewRequest("POST", targetUrl, payload)
		if err != nil {
			fmt.Println(err)
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+
			viper.GetString("glm-key"))

		response, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println(err)
		}
		defer response.Body.Close()          // 这步是必要的，防止以后的内存泄漏，切记
		body, _ := io.ReadAll(response.Body) // 读取响应 body, 返回为 []byte
		fmt.Println(string(body))            // 转成字符串看一下结果

		var res service.GlmResponse

		// 使用json.Unmarshal将字节数组解析到结构体中
		err = json.Unmarshal(body, &res)
		if err != nil {
			fmt.Println("解析JSON时出错:", err)
			return
		}

		fmt.Println(len(res.Choices))
		// wsc.sendMsgThread(res.Message[0])
		for i := 0; i < len(res.Choices); i++ {
			wsc.sendMsgThread(res.Choices[i].Message.Content)

			// 向请求参数的数组中添加消息
			reqData.Messages = service.ShiftTheMessages(reqData.Messages)
			reqData.Messages = append(reqData.Messages, res.Choices[i].Message)

			time.Sleep(time.Second * 2)
		}
	}()
}

func init() {
	err := setConfigInfo()
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
}
func main() {
	// 初始化请求参数
	reqData = NewRequest{
		Model:    "glm-4",
		Messages: []service.ReqMessage{},
	}

	wsc := NewWsClientManager("8.141.5.195", "9079", "/ws/"+viper.GetString("chat-uuid"), 10)
	wsc.start()

	var w1 sync.WaitGroup
	w1.Add(1)
	w1.Wait()
}

func setConfigInfo() error {
	// 读取启动参数
	flag.StringVar(&configPath, "c", "./config.yaml", "配置文件路径")
	flag.Parse()

	// 初始化配置文件
	viper.SetDefault("glm-key", "")
	viper.SetDefault("max-dialogue", 10)
	viper.SetDefault("chat-uuid", "ef1f53e2-5259-41b2-816b-7dbc0fddace8")

	// 启动参数割串
	arr := strings.Split(configPath, "/")

	var configPathStr string = ""
	for i := 0; i < len(arr)-1; i++ {
		configPathStr += arr[i]
		if i != len(arr)-2 {
			configPathStr += "/"
		}
	}
	configNameStr := strings.Split(arr[len(arr)-1], ".")[0]

	viper.SetConfigName(configNameStr)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPathStr)
	err := viper.ReadInConfig()
	return err
}
