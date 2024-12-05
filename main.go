package main

import (
	"context"
	"log"
	"viola/voc/internal"
)

func main() {
	// 初始化 bedrock client
	internal.InitBedrockClient()
	// 初始化数据库
	internal.InitDynamoDB()

	//从csv中插入数据
	//internal.InsertData()

	//从DB中读取数据
	str, err := internal.GetMsgFromDB()
	if err != nil {
		log.Fatal(err)
	}
	// 调用模型形成分类词
	result := internal.HandleBedrockClaude3SonnetV2(context.Background(), str)
	// 写入DB
	internal.InsertReviewResult(result)
}