package service

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/spf13/viper"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	hunyuan "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/hunyuan/v20230901"
)

func PostHunyuan(msg string, usr string) GlmResponse {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取
	credential := common.NewCredential(
		viper.GetString("secret-id"),
		viper.GetString("secret-key"),
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "hunyuan.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := hunyuan.NewClient(credential, "", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := hunyuan.NewChatCompletionsRequest()

	fmt.Println(viper.GetString("usr"))
	request.Model = common.StringPtr(viper.GetString("model"))
	request.Messages = []*hunyuan.Message{
		&hunyuan.Message{
			Role:    common.StringPtr("user"),
			Content: common.StringPtr(msg),
		},
	}
	request.Stream = common.BoolPtr(false)
	request.TopP = common.Float64Ptr(1)
	request.Temperature = common.Float64Ptr(1)

	// 返回的resp是一个ChatCompletionsResponse的实例，与请求对象对应
	response, err := client.ChatCompletions(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return GlmResponse{}
	}
	if err != nil {
		panic(err)
	}
	// 输出json格式的字符串回包
	if response.Response != nil {
		// 非流式响应
		fmt.Println(response.ToJsonString())
		// 定义一个包装结构体来解析 JSON 字符串
		type WrappedResponse struct {
			Response GlmResponse `json:"Response"`
		}

		// 解析 JSON 字符串为 WrappedResponse 结构体
		var wrappedResponse WrappedResponse
		err := json.Unmarshal([]byte(response.ToJsonString()), &wrappedResponse)
		if err != nil {
			log.Fatalf("Error unmarshalling JSON: %v", err)
		}

		// 获取解析后的 Response 结构体
		response := wrappedResponse.Response
		return response
		// res := GlmResponse{
		// 	Choices: response.Response.Choices,
		// }

	} else {
		// 流式响应
		for event := range response.Events {
			fmt.Println(string(event.Data))
		}
	}
	return GlmResponse{}
}
