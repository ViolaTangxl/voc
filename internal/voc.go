package internal

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"viola/voc/internal/model"

	"github.com/google/uuid"
)

// 上传csv文件的内容到DynamoDB
func handleUploadCSV(r *http.Request) ([]model.Review, error) {
	// 检查请求方法
	if r.Method != http.MethodPost {
		log.Print("Method not allowed")
		return nil, errors.New("method not allowed")
	}

	// 解析multipart form数据
	err := r.ParseMultipartForm(10 << 20) // 限制文件大小为10MB
	if err != nil {
		log.Print("Failed to parse form")
		return nil, errors.New("failed to parse form")
	}

	// 获取上传的文件
	file, _, err := r.FormFile("csv")
	if err != nil {
		log.Print("Failed to get file")
		return nil, errors.New("failed to get file")
	}
	defer file.Close()

	// 创建CSV reader
	reader := csv.NewReader(file)

	// 存储数据的临时slice
	var records []model.Review

	// 读取所有记录
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Print("Error reading CSV content")
			return nil, errors.New("error reading CSV content")
		}
		// 解析ReviewDate
		reviewDate, err := parseDateString(record[0])
		if err != nil {
			fmt.Println("Error parsing date:", err)
			continue
		}
		// 创建Review结构体
		review := model.Review{
			Id:            uuid.New().String(),
			ReviewDate:    reviewDate,
			ReviewContent: record[1],
			ProductType:   record[2], // 产品名称
		}

		records = append(records, review)
	}

	fmt.Printf("csv解析完成\n 长度：%d\n", len(records))
	return records, nil
}

// 调用模型开始分类
func tryToCategory(ctx context.Context, reviews []model.Review) {
	var combinedComments strings.Builder
	fmt.Println("----- 准备调用模型-----")
	// 模型分类
	for i, review := range reviews {
		if i == 0 {
			//	只有第一次才要写入商品名称
			combinedComments.WriteString(fmt.Sprintf("产品名称:%s\n", review.ProductType))
		}
		combinedComments.WriteString(fmt.Sprintf("评价: %s\n",
			review.ReviewContent))
	}
	// 调用模型
	resultStr := HandleBedrockClaude3SonnetV2(ctx, combinedComments.String())
	fmt.Println("----- 调用模型完成-----")

	// 解析结果拼装成ReviewResult
	var reviewResult model.ReviewResult
	err := json.Unmarshal([]byte(resultStr), &reviewResult)
	if err != nil {
		log.Fatalf("Error unmarshaling review result:", err)
	}
	// 产品名称赋值
	showProducts = append(showProducts, Item{
		ID:   reviewResult.ProductName,
		Name: reviewResult.ProductName,
	})

	// 产品一级分类
	firstCategory := make([]Item, 0)
	// 遍历每个一级分类
	for _, review := range reviewResult.Review {
		// 添加一级分类
		firstCategory = append(firstCategory, Item{
			ID:   review.FirstLevel,
			Name: review.FirstLevel,
		})

		// 为每个一级分类创建对应的二级分类数组
		secondCategory := make([]Item, 0)
		for _, sl := range review.SecondLevel {
			secondCategory = append(secondCategory, Item{
				ID:   sl.Categorization,
				Name: sl.Categorization,
			})
		}
		// 将当前一级分类对应的二级分类存储到 showSubCategories
		showSubCategories[review.FirstLevel] = secondCategory
	}

	// 存储产品对应的一级分类
	showCategories[reviewResult.ProductName] = firstCategory
	fmt.Printf("showCategories:【%+v】\n", showCategories)
	fmt.Printf("showSubCategories:【%+v】\n", showSubCategories)
	fmt.Println("----- 分类完成-----")
}
