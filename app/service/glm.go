package service

import (
	"beibei/app/common"

	"github.com/spf13/viper"
)

// glm响应
type GlmResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	FinishReason string            `json:"finish_reason"`
	Index        int               `json:"index"`
	Message      common.ReqMessage `json:"message"`
}

// 检查对话长度，若超过10句则删除前面的对话，使对话长度保持在10句
func ShiftTheMessages(msgs []common.ReqMessage) []common.ReqMessage {
	count := viper.GetInt("max-dialogue")

	if len(msgs) > count {
		msgs = msgs[len(msgs)-count:]
	}
	return msgs
}
