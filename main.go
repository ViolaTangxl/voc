package main

import (
	"viola/voc/internal"
)

func main() {
	// // 初始化 bedrock client
	internal.InitBedrockClient()
	// 初始化DynamoDB 数据库
	internal.InitDynamoDB()
	// 初始化Redshift 数据库
	internal.InitRedShift()

	// // 从csv中插入数据
	// // internal.InsertData()
	//
	// // 从DB中读取数据
	// str, err := internal.GetMsgFromDB()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// // 调用模型形成分类词
	// result := internal.HandleBedrockClaude3SonnetV2(context.Background(), str)
	// // 写入DB
	// internal.InsertReviewResult(result)
	internal.Start()
}
