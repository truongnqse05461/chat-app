package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"v2/aws"
	"v2/dao"
	"v2/routes"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error getting env, %v", err)
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(dir)
	dao.Init()
	router := gin.Default()
	//router := mux.NewRouter()
	//router.HandleFunc("/", routes.GetHome)
	//
	router.GET("/ws", routes.StartWebSocket)

	router.POST("/auth/register", routes.Register)
	router.POST("/auth/login", routes.Login)

	router.POST("/conv/clear", routes.Clear)

	router.POST("/user/clear", routes.ClearUser)

	router.GET("/conv/get", routes.GetMessages)

	router.GET("/user/info", routes.GetUserInfo)

	router.GET("/aws/upload", aws.UploadHandler)

	router.POST("/file/upload", routes.UploadFile)

	router.POST("/file/download", routes.DownloadFile)

	//fileServer := http.FileServer(http.Dir("./data"))

	router.Static("/static", "./data/")
	//router.StaticFile("/static/Capture.PNG", "./data/Capture.PNG")

	fmt.Println("Server starting at :8080")

	router.Run(":8080")

	//headersOk := handlers.AllowedHeaders([]string{"X-Requested-With"})
	//originsOk := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	//methodsOk := handlers.AllowedMethods([]string{"GET", "HEAD", "POST", "PUT", "OPTIONS"})
	//http.ListenAndServe(":8080", handlers.CORS(originsOk, headersOk, methodsOk)(router))
}
