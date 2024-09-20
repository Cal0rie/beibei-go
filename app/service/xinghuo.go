package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"beibei/app/common"

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

// 发送POST请求
func (wsc *websocketClientManager) XinghuoPost(msg string, usr string, reqData *common.NewRequest) {
	go func() {
		// 默认情况下为glm
		targetUrl := "https://spark-api-open.xf-yun.com/v1/chat/completions"
		fmt.Println(msg)
		// reqData := NewRequest{
		// 	Session_id: "friend-123",
		// 	Username:   usr,
		// 	Message:    msg,
		// }

		reqData.Messages = ShiftTheMessages(reqData.Messages)
		// 向请求参数的数组中添加消息
		reqData.Messages = append(reqData.Messages, common.ReqMessage{
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

		var res GlmResponse

		// 使用json.Unmarshal将字节数组解析到结构体中
		err = json.Unmarshal(body, &res)
		if err != nil {
			fmt.Println("解析JSON时出错:", err)
			return
		}

		fmt.Println(len(res.Choices))
		// wsc.sendMsgThread(res.Message[0])
		for i := 0; i < len(res.Choices); i++ {
			// wsc.sendMsgThread(res.Choices[i].Message.Content)

			// 向请求参数的数组中添加消息
			reqData.Messages = ShiftTheMessages(reqData.Messages)
			reqData.Messages = append(reqData.Messages, res.Choices[i].Message)

			time.Sleep(time.Second * 2)
		}
	}()
}
