package internal

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
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
