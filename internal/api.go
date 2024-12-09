package internal

import (
	"context"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
	"viola/voc/internal/model"

	"github.com/gin-gonic/gin"
)

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var mu sync.Mutex // 或者使用 sync.RWMutex

type SecondLevel struct {
	Categorization string `json:"categorization"`
}

type Comment struct {
	FirstLevel  string        `json:"first_level"`
	Percentage  string        `json:"percentage,omitempty"`
	SecondLevel []SecondLevel `json:"second_level"`
}

type Product struct {
	ProductName string    `json:"product_name"`
	Comments    []Comment `json:"comments"`
}

// 初始化数据
//
//	var data = ProductList{
//		Products: []Product{
//			{
//				ProductName: "修容盘",
//				Comments: []Comment{
//					{
//						FirstLevel: "产品问题",
//						Percentage: "75%",
//						SecondLevel: []SecondLevel{
//							{Categorization: "色彩表现"},
//							{Categorization: "持久度"},
//						},
//					},
//					{
//						FirstLevel: "物流问题",
//						Percentage: "10%",
//						SecondLevel: []SecondLevel{
//							{Categorization: "配送速度"},
//							{Categorization: "包装完整性"},
//						},
//					},
//					{
//						FirstLevel: "包装问题",
//						SecondLevel: []SecondLevel{
//							{Categorization: "外观设计"},
//							{Categorization: "质量问题"},
//						},
//					},
//				},
//			},
//			{
//				ProductName: "眼线笔",
//				Comments: []Comment{
//					{
//						FirstLevel: "123产品问题",
//						Percentage: "75%",
//						SecondLevel: []SecondLevel{
//							{Categorization: "123色彩表现"},
//							{Categorization: "123持久度"},
//						},
//					},
//					{
//						FirstLevel: "123物流问题",
//						Percentage: "10%",
//						SecondLevel: []SecondLevel{
//							{Categorization: "123配送速度"},
//							{Categorization: "123包装完整性"},
//						},
//					},
//					{
//						FirstLevel: "123包装问题",
//						SecondLevel: []SecondLevel{
//							{Categorization: "123外观设计"},
//							{Categorization: "123质量问题"},
//						},
//					},
//				}},
//		},
//	}
var ProductResp = make([]model.ReviewResult, 0)

func ProductsHandle(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	mu.Lock()
	defer mu.Unlock()

	for len(ProductResp) == 0 {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				c.JSON(http.StatusRequestTimeout, gin.H{"error": "timeout waiting for products"})
			} else {
				c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request cancelled"})
			}
			return
		default:
			mu.Unlock()
			time.Sleep(100 * time.Millisecond)
			mu.Lock()
		}
	}

	c.JSON(http.StatusOK, ProductResp)
}

//func CategoriesHandle(c *gin.Context) {
//	id := c.Param("id")
//	c.JSON(http.StatusOK, showCategories[id])
//}
//
//func SubcategoriesHandle(c *gin.Context) {
//	id := c.Param("id")
//	c.JSON(http.StatusOK, showSubCategories[id])
//}

func HandleUploadCSV(c *gin.Context) {
	// 构建参数，调用voc.go 中的handleUploadCSV
	ctx := c.Request.Context()
	results, err := handleUploadCSV(c.Request)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	go tryToCategory(context.Background(), results)

	// 将数据存入DynamoDB
	err = batchWriteToDynamoDB(ctx, results)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CSV uploaded successfully"})
}

// IndexHandle 首页展示
func IndexHandle(c *gin.Context) {
	filePath := "./index.html"
	content, err := os.ReadFile(filePath)
	if err != nil {
		log.Println(err)
		return
	}
	c.Data(200, "text/html; charset=utf-8", content)
	return
}

// 开始模型分类
func HandleBedrockCategory(c *gin.Context) {
	// 构建参数，调用voc.go 中的handleUploadCSV
}

func Start() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	api := r.Group("/api")
	api.GET("/index", IndexHandle)
	api.GET("/products", ProductsHandle)
	//api.GET("/categories/:id", CategoriesHandle)
	//api.GET("/subcategories/:id", SubcategoriesHandle)
	api.POST("/upload-csv", HandleUploadCSV)
	// 开始尝试分类
	api.POST("/try-to-category", HandleBedrockCategory)
	r.Run(":8080")
}
