package types

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Type struct {
	Name string `json:"name" binding:"required"`
}
type TypeResp struct {
	Id   int
	Name string
}

type TypeList struct {
	Types []TypeResp `json:"types" binding:"required"`
}

func main(){
	router := gin.Default()
	router.POST("/v1/type", createType)
	router.GET("/v1/type/:id", getType)
	router.GET("/v1/type/list", getTypeList)
	router.Run(":7000")
}

func getTypeList(c *gin.Context) {
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	fmt.Println(err)
	defer db.Close()
	rows, err := db.Query(`
    select id,name from types
    `)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	cateList := TypeList{}
	for rows.Next() {
		typ := TypeResp{}
		err = rows.Scan(&typ.Id, &typ.Name)
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		cateList.Types = append(cateList.Types, typ)
	}
	fmt.Println(cateList)
	c.JSON(http.StatusOK, cateList)
}

func getType(c *gin.Context) {
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
			"error": "wrong type for Type id",
		})
		return
	}
	var Type Type
	err = db.QueryRow(`
    select name from types where id=$1
    `, id).Scan(&Type.Name)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "wrong argument in getting Type name",
		})
		return
	}
	c.JSON(http.StatusOK, Type)

}

func createType(c *gin.Context){
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)

	}
	defer db.Close()
	var (
		typ Type
	)
	err = c.ShouldBindJSON(&typ)
	if err != nil {
		fmt.Println("error while binding type", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	var typeResp TypeResp
	err = db.QueryRow(`insert into types(name) values($1) returning id,name`, typ.Name).Scan(&typeResp.Id, &typeResp.Name)
	if err != nil {
		fmt.Println("error while inserting type", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	c.JSON(http.StatusCreated, typeResp)
}
