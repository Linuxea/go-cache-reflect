package cache

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var db *gorm.DB
var redisCli *redis.Client

func init() {
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			LogLevel: logger.Info,
		}),
	})
	if err != nil {
		panic("failed to connect database")
	}
}

func init() {
	redisCli = redis.NewClient(&redis.Options{
		PoolSize: 10,
		Addr:     "172.17.0.1:6379",
		Password: "",
		DB:       0,
	})
}

// func init() {
// 	type Product struct {
// 		gorm.Model
// 		Code  string
// 		Price uint
// 	}
// 	db.Create(&Product{Code: "D42", Price: 100, Model: gorm.Model{
// 		CreatedAt: time.Now(),
// 		UpdatedAt: time.Now(),
// 	}})
// }

func TestList(t *testing.T) {

	type Product struct {
		gorm.Model
		Code  string
		Price uint
	}

	db.AutoMigrate(&Product{})

	var result []*Product

	// list
	cache := &Cache{key: "productAll",
		redis: redisCli,
		model: &result,
		FetchFun: func() (interface{}, error) {
			list := make([]*Product, 0)
			return &list, db.Find(&list).Error
		},
		ttl: time.Second * 10,
	}
	err := cache.cache()
	if err != nil {
		panic(err)
	}

	for _, temp := range result {
		fmt.Printf("%d 价格对应为 %d\n", temp.ID, temp.Price)
	}

}

func TestGetString(t *testing.T) {
	// get string
	var result string
	getStringCache := &Cache{key: "randomString",
		redis: redisCli,
		model: &result,
		FetchFun: func() (interface{}, error) {
			s := time.Now().Format("2006-01-02 15:04:05")
			return &s, nil
		},
		ttl: time.Second * 10,
	}

	err := getStringCache.cache()
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}

func TestGetInt(t *testing.T) {
	// get int
	var result int
	getIntCache := &Cache{key: "3333" + "id",
		redis: redisCli,
		model: &result,
		FetchFun: func() (interface{}, error) {
			second := time.Now().Second()
			return &second, nil
		},
		ttl: time.Second * 10,
	}

	err := getIntCache.cache()
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}

func TestGetStruct(t *testing.T) {

	type Product struct {
		gorm.Model
		Code  string
		Price uint
	}

	var result Product

	// get struct
	getOneProductCache := &Cache{key: "oneProduct",
		redis: redisCli,
		model: &result,
		FetchFun: func() (interface{}, error) {
			return &Product{Code: "abcdefg", Price: 1000, Model: gorm.Model{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}}, nil
		},
		ttl: time.Second * 10,
	}

	err := getOneProductCache.cache()
	if err != nil {
		panic(err)
	}
	fmt.Println(result)
}

func TestGetComplicateStruct(t *testing.T) {

	type Product struct {
		gorm.Model
		Code  string
		Price uint
	}

	type Person struct {
		Products []*Product
		Name     string
		Age      int
	}

	var result Person

	// get Person
	getPersonCache := &Cache{key: "getPerson",
		redis: redisCli,
		model: &result,
		FetchFun: func() (interface{}, error) {
			return &Person{
				Products: []*Product{{Code: "abcdefg", Price: 1000, Model: gorm.Model{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}}},
				Name: "linuxea",
				Age:  12,
			}, nil
		},
		ttl: time.Second * 30,
	}

	err := getPersonCache.cache()
	if err != nil {
		panic(err)
	}

	fmt.Println(result)
}
