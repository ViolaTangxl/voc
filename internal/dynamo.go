package internal

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"viola/voc/internal/adapter"
	"viola/voc/internal/model"
	"viola/voc/utils"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
)

type Review struct {
	ID            string `json:"id"`
	ReviewDate    int64  `json:"review_date"`
	ReviewContent string `json:"review_content"`
	ProductType   string `json:"product_type"`
}

var (
	DBClient *dynamodb.DynamoDB
)

func InitDynamoDB() {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(utils.ACCESS_KEY, utils.SECRET_KEY, ""),
	})
	if err != nil {
		log.Fatalf("Error creating AWS session:%s", err)
	}

	DBClient = dynamodb.New(sess)
}

func InsertData() {
	// 打开CSV文件
	file, err := os.Open("液体修容高光.csv")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// 创建CSV读取器
	reader := csv.NewReader(file)
	reader.FieldsPerRecord = -1 // 允许不同行有不同数量的字段

	// 迭代CSV记录并导入到DynamoDB
	for {
		record, err := reader.Read()
		if err != nil {
			break
		}

		// 解析ReviewDate
		reviewDate, err := parseDateString(record[0])
		if err != nil {
			fmt.Println("Error parsing date:", err)
			continue
		}
		// 写个方法，生成uuid

		// 创建Review结构体
		review := Review{
			ID:            uuid.New().String(),
			ReviewDate:    reviewDate,
			ReviewContent: record[1],
			ProductType:   "液体修容高光",
		}

		// 将Review结构体转换为AWS DynamoDB映射的值
		av, err := dynamodbattribute.MarshalMap(review)
		if err != nil {
			fmt.Println("Error marshaling review:", err)
			continue
		}

		// 将数据插入DynamoDB表
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("judydoll_product_review"), // 替换为您的DynamoDB表名
		}

		_, err = DBClient.PutItem(input)
		if err != nil {
			fmt.Println("Error inserting review:", err)
			continue
		}

		fmt.Println("Review inserted successfully")
	}
}

func parseDateString(dateStr string) (int64, error) {
	// 解析自然语言日期字符串
	dateStr = strings.ToLower(dateStr)
	dateStr = strings.TrimSpace(dateStr)

	var days int
	var err error
	if strings.HasSuffix(dateStr, "天前") {
		days, err = strconv.Atoi(strings.TrimSuffix(dateStr, "天前"))
	} else if strings.HasSuffix(dateStr, "个月前") {
		months, err := strconv.Atoi(strings.TrimSuffix(dateStr, "个月前"))
		if err == nil {
			days = months * 30
		}
	} else {
		return 0, fmt.Errorf("invalid date string: %s", dateStr)
	}

	if err != nil {
		return 0, err
	}

	// 计算时间戳
	timestamp := time.Now().AddDate(0, 0, -days).Unix()
	return timestamp, nil
}

func GetMsgFromDB() (string, error) {
	// 从DynamoDB表中获取数据
	var combinedComments strings.Builder
	for _, productType := range utils.ProductTypeList {
		input := &dynamodb.ScanInput{
			TableName:        aws.String("judydoll_product_review"),
			FilterExpression: aws.String("product_type = :product_type"),
			ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
				":product_type": {
					S: aws.String(productType),
				},
			},
		}
		// 获取数据
		result, err := DBClient.Scan(input)
		if err != nil {
			fmt.Println("Error scanning table:", err)
			return "", err
		}
		// 解析数据
		var reviews []Review
		err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &reviews)
		fmt.Println(len(reviews))
		combinedComments.WriteString(fmt.Sprintf("产品名称:%s\n", productType))
		for _, review := range reviews {
			combinedComments.WriteString(fmt.Sprintf("评价: %s\n",
				review.ReviewContent))
		}
		// TODO 这里只取一种商品
		break
	}

	return combinedComments.String(), nil
}

// 写个方法，传入的是json字符串，然后解析成ReviewResult，存入DynamoDB
func InsertReviewResult(reviewResultStr string) {
	var reviewResult model.ReviewResult
	err := json.Unmarshal([]byte(reviewResultStr), &reviewResult)
	if err != nil {
		log.Fatalf("Error unmarshaling review result:", err)
	}
	for _, review := range adapter.ConvertReviewResultToReviewResultDBModel(reviewResult) {
		// 将Review结构体转换为AWS DynamoDB映射的值
		av, err := dynamodbattribute.MarshalMap(review)
		if err != nil {
			log.Fatalf("Error marshaling review:", err)
		}

		// 将数据插入DynamoDB表
		input := &dynamodb.PutItemInput{
			Item:      av,
			TableName: aws.String("judydoll_product_review_result"), // 替换为您的DynamoDB表名
		}

		_, err = DBClient.PutItem(input)
		if err != nil {
			log.Fatalf("Error inserting review:", err)
		}

		fmt.Println("Review inserted successfully")
	}

	fmt.Println("ALL review inserted successfully")
}
