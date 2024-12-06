package internal

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"os"
)

type Item struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var products = []Item{
	{ID: "1", Name: "手机"},
	{ID: "2", Name: "电脑"},
	{ID: "3", Name: "家电"},
}

var categories = map[string][]Item{
	"1": {{ID: "11", Name: "苹果"}, {ID: "12", Name: "华为"}, {ID: "13", Name: "小米"}},
	"2": {{ID: "21", Name: "笔记本"}, {ID: "22", Name: "台式机"}, {ID: "23", Name: "平板"}},
	"3": {{ID: "31", Name: "冰箱"}, {ID: "32", Name: "洗衣机"}, {ID: "33", Name: "空调"}},
}

var subCategories = map[string][]Item{
	"11": {{ID: "111", Name: "iPhone"}, {ID: "112", Name: "iPad"}},
	"12": {{ID: "121", Name: "Mate系列"}, {ID: "122", Name: "P系列"}},
	"13": {{ID: "131", Name: "小米手机"}, {ID: "132", Name: "红米手机"}},
	"21": {{ID: "211", Name: "游戏本"}, {ID: "212", Name: "商务本"}},
	"22": {{ID: "221", Name: "一体机"}, {ID: "222", Name: "组装机"}},
	"23": {{ID: "231", Name: "Android平板"}, {ID: "232", Name: "Windows平板"}},
	"31": {{ID: "311", Name: "对开门"}, {ID: "312", Name: "十字对开门"}},
	"32": {{ID: "321", Name: "滚筒"}, {ID: "322", Name: "波轮"}},
	"33": {{ID: "331", Name: "挂机"}, {ID: "332", Name: "柜机"}},
}

func ProductsHandle(c *gin.Context) {
	c.JSON(http.StatusOK, products)
}

func CategoriesHandle(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, categories[id])
}

func SubcategoriesHandle(c *gin.Context) {
	id := c.Param("id")
	c.JSON(http.StatusOK, subCategories[id])
}

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

func Start() {
	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	api := r.Group("/api")
	api.GET("/index", IndexHandle)
	api.GET("/products", ProductsHandle)
	api.GET("/categories/:id", CategoriesHandle)
	api.GET("/subcategories/:id", SubcategoriesHandle)

	r.Run(":8080")
}
