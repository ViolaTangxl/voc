package model

type Review struct {
	Id            string `json:"id" dynamodbav:"id"`             // Added dynamodbav tag
	ReviewDate    int64  `json:"review_date" dynamodbav:"review_date"`
	ReviewContent string `json:"review_content" dynamodbav:"review_content"`
	ProductType   string `json:"product_type" dynamodbav:"product_type"`
}
