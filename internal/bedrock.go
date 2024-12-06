package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"viola/voc/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/bedrockagentruntime"
	"github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
	"github.com/tidwall/gjson"
)

// bedrock runtime client
var BedrockClient *bedrockruntime.Client

// bedrock agent runtime client
var BedrockAgentRuntimeClient *bedrockagentruntime.Client

func InitBedrockClient() {
	// 创建凭证提供程序函数
	provider := aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
		return aws.Credentials{
			AccessKeyID:     utils.ACCESS_KEY,
			SecretAccessKey: utils.SECRET_KEY,
		}, nil
	})

	// load aws credentials from profile demo using config
	awsCfg1, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(utils.BEDROCK_REGION),
		config.WithCredentialsProvider(provider),
	)
	if err != nil {
		log.Fatal(err)
	}

	// create bedrock runtime client
	BedrockClient = bedrockruntime.NewFromConfig(awsCfg1)

	// create bedrock agent runtime client
	BedrockAgentRuntimeClient = bedrockagentruntime.NewFromConfig(awsCfg1)
}

// claude3 request data type
type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Message struct {
	Role    string    `json:"role"`
	Content []Content `json:"content"`
}

//	type Content struct {
//		Type string `json:"type"`
//		Text string `json:"text"`
//	}
//
//	type Message struct {
//		Role    string    `json:"role"`
//		Content []Content `json:"content"`
//	}
func HandleBedrockClaude3Haiku(ctx context.Context, comments string) {
	// Create request
	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
		ContentType: aws.String("application/json"),
	}

	// Create request body
	requestBody := struct {
		Messages         []Message `json:"messages"`
		MaxTokens        int       `json:"max_tokens"`
		Temperature      float64   `json:"temperature"`
		TopP             float64   `json:"top_p"`
		AnthropicVersion string    `json:"anthropic_version"`
		SystemPrompt     string    `json:"system"`
	}{
		Messages: []Message{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: comments,
					},
				},
			},
		},
		MaxTokens:        200000,
		Temperature:      0.7,
		TopP:             0.9,
		AnthropicVersion: "bedrock-2023-05-31",
		SystemPrompt: "你是一位高级数据分析师，你可以根据收集上来的各种语言的用户评价信息，将信息进行聚合分类。" +
			"请针对每次的用户评价，输出简体中文版的分类词" +
			"输出一份评价分类词，如一级分类：物流问题，二级分类：配送不及时、不送货上门等等；如一级分类：产品问题，二级分" +
			"类是每种产品项的具体问题分类，如不保湿、易结块等等。并确保至少有2个一级分类，不同二级分类的内容不能重叠或者类似。" +
			"一定要有物流问题这个一级分类。" +
			//"请每种产品分开分析总结。" +
			"如果有用户评价出现分歧，请总结百分比例。" +
			"不需要有正面评价、负面评价这种分类，",
		//"物流快递相关信息不要遗漏,一定要有物流快递相关一级分类" +
	}

	// Serialize request body
	var err error
	input.Body, err = json.Marshal(requestBody)
	if err != nil {
		log.Fatal(err)
	}
	// Call model
	output, err := BedrockClient.InvokeModel(ctx, input)
	if err != nil {
		log.Fatal(err)
	}

	// Get just the text from the response
	text := gjson.Get(string(output.Body), "content.0.text").String()
	// Print result
	log.Println(text)
	log.Println("=========")
	log.Println(string(output.Body))
	// {"id":"msg_bdrk_01CxB7ecerAPSPL4ViGJWviA","type":"message","role":"assistant","model":"claude-3-haiku-20240307","content":[{"type":"text","text":"根据收集的用户评价信息,可以总结出以下分类:\n\n一级分类:\n1. 产品问题\n2. 正面评价\n\n二级分类:\n产品问题:\n1. 不卡粉\n2. 不浮粉\n3. 提亮效果差\n4. 颜色不自然\n5. 难推开\n\n正面评价:\n1. 质地细腻\n2. 颜色自然\n3. 持久情况好\n4. 上妆效果好\n5. 适合亚洲人肤质\n\n总结如下:\n1. 液体修容高光:\n一级分类:产品问题(30%)、正面评价(70%)\n二级分类:\n产品问题:不卡粉(10%)、不浮粉(10%)、提亮效果差(10%)\n正面评价:质地细腻(20%)、颜色自然(20%)、持久情况好(10%)、上妆效果好(10%)、适合亚洲人肤质(10%)\n\n2. 修容盘:\n一级分类:产品问题(20%)、正面评价(80%)\n二级分类:\n产品问题:颜色不自然(10%)、难推开(10%)\n正面评价:质地细腻(20%)、颜色自然(20%)、持久情况好(10%)、上妆效果好(20%)、适合亚洲人肤质(10%)"}],"stop_reason":"end_turn","stop_sequence":null,"usage":{"input_tokens":84750,"output_tokens":409}}
	inputToken := gjson.Get(string(output.Body), "usage.input_tokens").Int()
	log.Printf("input token: %d", inputToken)
	outputToken := gjson.Get(string(output.Body), "usage.output_tokens").Int()
	log.Printf("output token: %d", outputToken)
}

//func HandleBedrockClaude3Haiku(ctx context.Context, comments string) {
//	// Create request
//	input := &bedrockruntime.InvokeModelInput{
//		ModelId:     aws.String("anthropic.claude-3-haiku-20240307-v1:0"),
//		ContentType: aws.String("application/json"),
//	}
//
//	// Create request body
//	requestBody := struct {
//		Prompt            string  `json:"prompt"`
//		MaxTokensToSample int     `json:"max_tokens_to_sample"`
//		Temperature       float64 `json:"temperature"`
//		TopP              float64 `json:"top_p"`
//	}{
//		Prompt:            comments,
//		MaxTokensToSample: 2048,
//		Temperature:       0.7,
//		TopP:              0.7,
//	}
//
//	// Serialize request body
//	var err error
//	input.Body, err = json.Marshal(requestBody)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Call model
//	output, err := BedrockClient.InvokeModel(ctx, input)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Print result
//	log.Println(string(output.Body))
//}

func HandleBedrockClaude3SonnetV2(ctx context.Context, comments string) string {
	// Create request

	// Use the inference profile ARN instead of direct model ID
	//inferenceProfileArn := "arn:aws:bedrock:us-east-1:ACCOUNT_ID:inference-profile/PROFILE_NAME"

	input := &bedrockruntime.InvokeModelInput{
		ModelId:     aws.String(utils.CLAUDE_35_SONNETV2),
		ContentType: aws.String("application/json"),
	}

	// Create request body
	requestBody := struct {
		Messages         []Message `json:"messages"`
		MaxTokens        int       `json:"max_tokens"`
		Temperature      float64   `json:"temperature"`
		TopP             float64   `json:"top_p"`
		AnthropicVersion string    `json:"anthropic_version"`
		SystemPrompt     string    `json:"system"`
	}{
		Messages: []Message{
			{
				Role: "user",
				Content: []Content{
					{
						Type: "text",
						Text: comments,
					},
				},
			},
		},
		MaxTokens:        200000,
		Temperature:      1,
		TopP:             0.999,
		AnthropicVersion: utils.CLAUDE_35_ANTHROPICVERSION,
		SystemPrompt: "你是一位高级数据分析师，你可以根据收集上来的各种语言的用户评价信息，将信息进行聚合分类。" +
			"请针对每次的用户评价，输出简体中文版的分类词" +
			"输出一份评价分类词，如一级分类：物流问题，二级分类：配送不及时、不送货上门等等；如一级分类：产品问题，二级分" +
			"类是每种产品项的具体问题分类，如保湿效果好、易结块等等。并确保至少有2个一级分类，不同二级分类的内容不能重叠或者类似。" +
			"一定要有物流问题这个一级分类。" +
			"产品效果、产品质地等都属于产品问题，可以统一归类,尽可能多的丰富二级分类" +
			"请总结百分比例。如果是没有评价内容的，可以直接忽略不参与统计" +
			"产品优点和缺点，都需要正确分类,但不需要总结。" +
			"结果以json的形式输出,如:{\"product_name\":\"阴影盘\",\n\"comments\":[{\n\"first_level\":\"物流问题\",\n\"percentage\":\"10%\",\n\"second_level\":[{\"categorization\":\"不送货上门\",\"percentage\":\"1%\",\"details\":[]}]}]}",
	}

	// Serialize request body
	var err error
	input.Body, err = json.Marshal(requestBody)
	if err != nil {
		fmt.Println("------------")
		fmt.Printf("err:%s\n", err)
		log.Fatal(err)
	}
	// Call model
	output, err := BedrockClient.InvokeModel(ctx, input)
	if err != nil {
		fmt.Println("========")
		fmt.Printf("err:%s\n", err)
		log.Fatal(err)
	}

	// Get just the text from the response
	text := gjson.Get(string(output.Body), "content.0.text").String()
	// Print result
	log.Println(text)
	// {"id":"msg_bdrk_01CxB7ecerAPSPL4ViGJWviA","type":"message","role":"assistant","model":"claude-3-haiku-20240307","content":[{"type":"text","text":"根据收集的用户评价信息,可以总结出以下分类:\n\n一级分类:\n1. 产品问题\n2. 正面评价\n\n二级分类:\n产品问题:\n1. 不卡粉\n2. 不浮粉\n3. 提亮效果差\n4. 颜色不自然\n5. 难推开\n\n正面评价:\n1. 质地细腻\n2. 颜色自然\n3. 持久情况好\n4. 上妆效果好\n5. 适合亚洲人肤质\n\n总结如下:\n1. 液体修容高光:\n一级分类:产品问题(30%)、正面评价(70%)\n二级分类:\n产品问题:不卡粉(10%)、不浮粉(10%)、提亮效果差(10%)\n正面评价:质地细腻(20%)、颜色自然(20%)、持久情况好(10%)、上妆效果好(10%)、适合亚洲人肤质(10%)\n\n2. 修容盘:\n一级分类:产品问题(20%)、正面评价(80%)\n二级分类:\n产品问题:颜色不自然(10%)、难推开(10%)\n正面评价:质地细腻(20%)、颜色自然(20%)、持久情况好(10%)、上妆效果好(20%)、适合亚洲人肤质(10%)"}],"stop_reason":"end_turn","stop_sequence":null,"usage":{"input_tokens":84750,"output_tokens":409}}
	inputToken := gjson.Get(string(output.Body), "usage.input_tokens").Int()
	log.Printf("input token: %d", inputToken)
	outputToken := gjson.Get(string(output.Body), "usage.output_tokens").Int()
	log.Printf("output token: %d", outputToken)

	return text
}
