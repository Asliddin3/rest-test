package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)



type Product struct {
	Name       string    `json:"name" binding:"required"`
	CategoryId int       `json:"categoryid" binding:"required"`
	TypeId     int       `json:"typeid" binding:"required"`
	Storage    []Storage `json:"stores" binding:"required"`
}
type ProductResp struct {
	Id         int
	Name       string    `json:"name" binding:"required"`
	CategoryId int       `json:"categoryid" binding:"required"`
	TypeId     int       `json:"typeid" binding:"required"`
	Storage    []Storage `json:"stores" binding:"required"`
}
type ProductInfo struct {
	Id       int
	Name     string
	Category string
	Type     string
}

type Storage struct {
	Id   int
	Name string `json:"name" binding:"required"`
	// Addresses []Address `json:"addresses" binding:"required"`
}

// type Address struct {
// 	Id       int
// 	District string `json:"district" binding:"required"`
// 	Street   string `json:"street" binding:"required"`
// }
type StorageInfo struct {
	Id   int
	Name string
}
type Category struct {
	Id   string `json:"id" binding:"required"`
	Name string `json:"name" binding:"required"`
}
type CategoryResp struct {
	Id   int
	Name string
}
type CategoryList struct {
	Categories []CategoryResp `json:"categories" binding:"required"`
}

func getCategories() (CategoryList, error) {
	client := resty.New()
	var p CategoryList
	_, err := client.R().SetResult(&p).Get("http://localhost:5000/v1/category/list")
	// fmt.Println(row)
	fmt.Println("in get category", p, err)
	if err != nil {
		log.Println("unable to connect Category")
		return CategoryList{}, err
	}
	return p, nil
}

type TypeResp struct {
	Id   int
	Name string
}

type TypeList struct {
	Types []TypeResp `json:"types" binding:"required"`
}

func getTypes() (TypeList, error) {
	client := resty.New()
	var p TypeList
	_, err := client.R().SetResult(&p).Get("http://localhost:7000/v1/type/list")
	// fmt.Println(row)
	fmt.Println("in get type", p, err)
	if err != nil {
		log.Println("unable to connect type")
		return TypeList{}, err
	}
	return p, nil
}

func main() {
	router := gin.Default()
	router.POST("/v1/product", createProduct)
	router.POST("/v1/products", createProducts)
	router.GET("/v1/product/info", getProduct)
	router.GET("/v1/category/list", func(c *gin.Context) {
		categrory, err := getCategories()
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		c.JSON(200, categrory)
	})
	router.GET("/v1/type/list", func(c *gin.Context) {
		types, err := getTypes()
		if err != nil {
			c.JSON(http.StatusBadRequest, err)
			return
		}
		c.JSON(200, types)
	})
	router.DELETE("v1/product/delete", deleteProduct)
	router.Run(":6000")
}

func deleteProduct(c *gin.Context) {
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)

	}
	defer db.Close()
	idstr := c.Request.URL.Query().Get("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		fmt.Println("error while converting id", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	_, err = db.Exec(`delete from products where id=$1`, id)
	if err != nil {
		fmt.Println("error while deleleting products", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, "product deleted")
}

func getProduct(c *gin.Context) {
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)

	}
	defer db.Close()
	idstr := c.Request.URL.Query().Get("id")
	id, err := strconv.Atoi(idstr)
	if err != nil {
		fmt.Println("error while converting id", err)
	}
	fmt.Println(id)
	row, err := db.Query(`select
	p.id,
	p.name,
	c.name,
	t.name
	FROM products p
	INNER JOIN categories c ON c.id=p.category_id
	INNER JOIN types t ON t.id=p.type_id
	WHERE p.id=$1`, id)
	if err != nil {
		fmt.Println("error while geting from row product")
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var products []ProductInfo
	for row.Next() {
		var product ProductInfo
		err = row.Scan(&product.Id, &product.Name, &product.Category, &product.Type)
		if err != nil {
			fmt.Println("error while scanning row products", err)
			c.JSON(http.StatusBadRequest, err)
			return
		}
		products = append(products, product)
	}
	if err != nil {
		fmt.Println("error while selecting product", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusFound, products)
}
func createProducts(c *gin.Context) {
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)

	}
	defer db.Close()
	tr, err := db.Begin()
	if err != nil {
		fmt.Println("error while begining transaction", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var products []Product
	err = c.ShouldBindJSON(&products)
	if err != nil {
		fmt.Println("error while binding json", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	var productsResp []ProductResp
	for _, product := range products {
		var productResp ProductResp
		err = tr.QueryRow(`insert into products(name,category_id,type_id) values($1,$2,$3) returning id,name,category_id,type_id`, product.Name, product.CategoryId, product.TypeId).Scan(&productResp.Id, &productResp.Name, &productResp.CategoryId, &productResp.TypeId)
		if err != nil {
			fmt.Println("error while inserting products", err)
			c.JSON(http.StatusBadRequest, err.Error())
			tr.Rollback()
			return
		}
		var storagesResp []Storage
		for _, storage := range product.Storage {
			var storageResp Storage
			err = tr.QueryRow(`insert into storages(name)values($1) returning id,name`, storage.Name).Scan(&storageResp.Id, &storageResp.Name)
			if err != nil {
				fmt.Println("error while inserting storages", err)
				c.JSON(http.StatusBadRequest, err.Error())
				tr.Rollback()
				return
			}
			_, err = tr.Exec(`insert into product_storages(storage_id,product_id) values($1,$2)`, storageResp.Id, productResp.Id)
			if err != nil {
				fmt.Println("error while inserting product_storages", err)
				c.JSON(http.StatusBadRequest, err)
				tr.Rollback()
				return
			}
			// var addressesResp []Address
			// for _, address := range storage.Addresses {
			// 	var addressResp Address
			// 	err = tr.QueryRow(`insert into addresses(district,street) values($1,$2) returning id,district,street`, address.District, address.Street).Scan(&addressResp.Id, &addressResp.District, &addressResp.Street)
			// 	if err != nil {
			// 		fmt.Println("error while inserting addresses", err)
			// 		c.JSON(http.StatusBadRequest, err)
			// 		tr.Rollback()
			// 		return
			// 	}
			// 	_, err = tr.Exec(`insert into storage_addresses(storage_id,address_id) values($1,$2)`, storageResp.Id, addressResp.Id)
			// 	if err != nil {
			// 		fmt.Println("error while inserting storage_addresses", err)
			// 		c.JSON(http.StatusBadRequest, err)
			// 		tr.Rollback()
			// 		return
			// 	}
			// 	addressesResp = append(addressesResp, addressResp)
			// storageResp.Addresses = addressesResp
			storagesResp = append(storagesResp, storageResp)
		}
		productResp.Storage = storagesResp
		productsResp = append(productsResp, productResp)
	}
	c.JSON(http.StatusCreated, productsResp)
}
func createProduct(c *gin.Context) {
	connStr := "user=postgres password=compos1995 dbname=productdb sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)

	}
	defer db.Close()
	tr, err := db.Begin()
	if err != nil {
		fmt.Println("error while begining transaction", err)
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var product Product
	err = c.ShouldBindJSON(&product)
	if err != nil {
		fmt.Println("error while binding json", err)
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	var productResp ProductResp
	err = tr.QueryRow(`insert into products(name,category_id,type_id) values($1,$2,$3) returning id,name,category_id,type_id`, product.Name, product.CategoryId, product.TypeId).Scan(&productResp.Id, &productResp.Name, &productResp.CategoryId, &productResp.TypeId)
	if err != nil {
		fmt.Println("error while inserting products", err)
		c.JSON(http.StatusBadRequest, err.Error())
		tr.Rollback()
		return
	}
	var storagesResp []Storage
	for _, storage := range product.Storage {
		var storageResp Storage
		err = tr.QueryRow(`insert into storages(name)values($1) returning id,name`, storage.Name).Scan(&storageResp.Id, &storageResp.Name)
		if err != nil {
			fmt.Println("error while inserting storages", err)
			c.JSON(http.StatusBadRequest, err.Error())
			tr.Rollback()
			return
		}
		_, err = tr.Exec(`insert into product_storages(storage_id,product_id) values($1,$2)`, storageResp.Id, productResp.Id)
		if err != nil {
			fmt.Println("error while inserting product_storages", err)
			c.JSON(http.StatusBadRequest, err)
			tr.Rollback()
			return
		}
		// var addressesResp []Address
		// for _, address := range storage.Addresses {
		// 	var addressResp Address
		// 	err = tr.QueryRow(`insert into addresses(district,street) values($1,$2) returning id,district,street`, address.District, address.Street).Scan(&addressResp.Id, &addressResp.District, &addressResp.Street)
		// 	if err != nil {
		// 		fmt.Println("error while inserting addresses", err)
		// 		c.JSON(http.StatusBadRequest, err)
		// 		tr.Rollback()
		// 		return
		// 	}
		// 	_, err = tr.Exec(`insert into storage_addresses(storage_id,address_id) values($1,$2)`, storageResp.Id, addressResp.Id)
		// 	if err != nil {
		// 		fmt.Println("error while inserting storage_addresses", err)
		// 		c.JSON(http.StatusBadRequest, err)
		// 		tr.Rollback()
		// 		return
		// 	}
		// 	addressesResp = append(addressesResp, addressResp)
		// }
		// storageResp.Addresses = addressesResp
		storagesResp = append(storagesResp, storageResp)
	}
	productResp.Storage = storagesResp
	c.JSON(http.StatusCreated, productResp)
}
