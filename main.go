package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"google.golang.org/appengine"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

func init() {
	log.Println("Hello, init!")
}

/*
	Main
*/
func main() {
	router := gin.Default()
	router.LoadHTMLGlob("./template/*.html")

	// Middleware
	//router.Use(RecordUaAndTime)

	ua := ""
	router.Use(func(c *gin.Context) {
		ua = c.GetHeader("User-Agent")
		c.Next()
	})

	//Index
	router.GET("/", func(ctx *gin.Context) {
		todos := dbGetAll()
		log.Printf("todos: %+v", todos)
		ctx.HTML(200, "index.html", gin.H{
			"todos": todos,
		})
	})

	//Create
	router.POST("/new", func(ctx *gin.Context) {
		text := ctx.PostForm("text")
		status := ctx.PostForm("status")
		dbInsert(text, status)
		ctx.Redirect(302, "/")
	})

	//Detail
	router.GET("/detail/:id", func(ctx *gin.Context) {
		n := ctx.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic(err)
		}
		todo := dbGetOne(id)
		ctx.HTML(200, "detail.html", gin.H{"todo": todo})
	})

	//Update
	router.POST("/update/:id", func(ctx *gin.Context) {
		n := ctx.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic("ERROR")
		}
		text := ctx.PostForm("text")
		status := ctx.PostForm("status")
		dbUpdate(id, text, status)
		ctx.Redirect(302, "/")
	})

	//削除確認
	router.GET("/delete_check/:id", func(ctx *gin.Context) {
		n := ctx.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic("ERROR")
		}
		todo := dbGetOne(id)
		ctx.HTML(200, "delete.html", gin.H{"todo": todo})
	})

	//Delete
	router.POST("/delete/:id", func(ctx *gin.Context) {
		n := ctx.Param("id")
		id, err := strconv.Atoi(n)
		if err != nil {
			panic("ERROR")
		}
		dbDelete(id)
		ctx.Redirect(302, "/")

	})

	/*
		data := "Hello Go/Gin!!"

		router.GET("/", func(ctx *gin.Context) {
			ctx.HTML(200, "index.html", gin.H{"data": data})
		})

		router.GET("/test", func(ctx *gin.Context) {
			ctx.JSON(http.StatusOK, gin.H{
				"data": data,
				"ua":   ua,
			})
		})
	*/

	// appengine用
	// router.Run()の代わり
	//router.Run()
	http.Handle("/", router)
	appengine.Main()
}

func index(w http.ResponseWriter, r *http.Request) {
	log.Println("Hello, world!")
	json.NewEncoder(w).Encode(Response{Status: "ok", Message: "Hello world!"})
}

func myError(err error) error {
	_, file, line, _ := runtime.Caller(1)
	newErr := fmt.Errorf("[ERROR] %+v:%+v %w", file, line, err)
	return newErr
}

type Todo struct {
	gorm.Model
	Text   string
	Status string
}

/*
 Database
*/
func dbOpen() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil || db == nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Todo{})
	sqlDB, _ := db.DB()

	// SetMaxIdleConns sets the maximum number of connections in the idle connection pool.
	sqlDB.SetMaxIdleConns(10)

	// SetMaxOpenConns sets the maximum number of open connections to the database.
	sqlDB.SetMaxOpenConns(100)

	// SetConnMaxLifetime sets the maximum amount of time a connection may be reused.
	sqlDB.SetConnMaxLifetime(time.Hour)

	return db
}

func dbGetAll() []Todo {

	var todos []Todo

	db := dbOpen()
	if db != nil {
		db.Order("created_at desc").Find(&todos)
	}
	return todos
}

func dbInsert(text string, status string) {
	db := dbOpen()
	db.Create(&Todo{Text: text, Status: status})
}

func dbGetOne(id int) Todo {
	var todo Todo
	db := dbOpen()
	db.First(&todo, id)
	return todo
}

func dbUpdate(id int, text string, status string) {
	db := dbOpen()

	var todo Todo
	db.First(&todo, id)
	todo.Text = text
	todo.Status = status
	db.Save(&todo)
}

func dbDelete(id int) {
	db := dbOpen()

	var todo Todo
	db.First(&todo, id)
	db.Delete(&todo)
}
