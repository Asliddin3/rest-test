package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// type PrintJob struct {
// 	Format    string `json:"format" binding:"required"`
// 	InvoiceId int    `json:"invoiceId" binding:"required,gte=0"`
// 	JobId     int    `json:"jobId" binding:"gte=0"`
// }

type Category struct {
	Name string `json:"name" binding:"required"`
}
type CategoryResp struct {
	Id   int
	Name string
}
type CategoryList struct {
	Categories []CategoryResp `json:"categories" binding:"required"`
}

func main() {
	router := gin.Default()
	// router.POST("/category", func(c *gin.Context) {
	// 	var p Category
	// 	if err := c.ShouldBindJSON(&p); err != nil {
	// 		c.JSON(400, gin.H{"error": "Invalid input!"})
	// 		return
	// 	}
	// 	// log.Printf("PrintService: creating new print job from invoice #%v...", p.InvoiceId)
	// 	// rand.Seed(time.Now().UnixNano())
	// 	// p.JobId = rand.Intn(1000)
	// 	// log.Printf("PrintService: created print job #%v", p.JobId)
	// 	createCategory(c)
	// 	c.JSON(200, p)
	// })
	router.POST("/v1/categories", createCategories)
	router.POST("/v1/category", createCategory)
	router.GET("/v1/category/:id", getCategory)
	router.GET("/v1/category/list", getCategoryList)
	router.Run(":5000")
}

func getCategoryList(c *gin.Context) {
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	fmt.Println(err)
	defer db.Close()
	rows, err := db.Query(`
    select id,name from categories
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	cateList := CategoryList{}
	for rows.Next() {
		cate := CategoryResp{}
		err = rows.Scan(&cate.Id, &cate.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		cateList.Categories = append(cateList.Categories, cate)
	}
	fmt.Println(cateList)
	c.JSON(http.StatusOK, cateList)
}

func getCategory(c *gin.Context) {
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	idstr := c.Param("id")
	fmt.Println(idstr)
	id, err := strconv.Atoi(idstr)
	fmt.Println(err == nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong type for category id",
		})
		return
	}
	var category Category
	err = db.QueryRow(`
    select name from categories where id=$1
    `, id).Scan(&category.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong argument in getting category name",
		})
		return
	}
	c.JSON(http.StatusOK, category)

}

func createCategories(c *gin.Context) {
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)

	}
	defer db.Close()
	var categories []Category
	err = c.ShouldBindJSON(&categories)
	if err != nil {
		fmt.Println("error while binding categories", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	var categorieResp []CategoryResp
	for _, val := range categories {
		var cateResp CategoryResp
		err = db.QueryRow(`insert into categories(name) values($1) returning id,name`, val.Name).Scan(&cateResp.Id, &cateResp.Name)
		if err != nil {
			fmt.Println(err)
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		categorieResp = append(categorieResp, cateResp)
	}
	c.JSON(http.StatusCreated, categorieResp)
}

func createCategory(c *gin.Context) {
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	fmt.Println("error in db", err)
	if err != nil {
		panic(err)

	}
	defer db.Close()
	var (
		category Category
	)
	err = c.ShouldBindJSON(&category)
	if err != nil {
		fmt.Println("error while binding", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	var cateRes CategoryResp
	err = db.QueryRow(`insert into categories(name) values($1) returning id,name`, category.Name).Scan(&cateRes.Id, &cateRes.Name)
	if err != nil {
		fmt.Println(err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, cateRes)
}
