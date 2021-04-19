package routes

import (
	"context"
	"fmt"
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
	"v2/db"
	"v2/models"
	"v2/utils"
)

var database = db.MongoConnection()

type FileMessage struct {
	FilePath string
	Content  string
}

func Register(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	//upload size
	r.ParseMultipartForm(200000)

	//reading original file
	file, handler, errReadFile := r.FormFile("originalFile")
	if errReadFile != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(errReadFile)
		return
	}
	defer file.Close()

	fileType := strings.Split(handler.Filename, ".")[1]
	fileName := uuid.NewString() + "." + fileType
	resFile, errCreateFile := os.Create("./data/" + fileName)
	if errCreateFile != nil {
		fmt.Fprintln(w, errCreateFile)
	}
	defer resFile.Close()
	if errCreateFile == nil {
		io.Copy(resFile, file)
		defer resFile.Close()
	}

	if govalidator.IsNull(username) || govalidator.IsNull(password) {

		utils.JSONResponse(w, 400, nil, 0, "Bad request - data can not empty")
		return
	}
	userCollection := db.GetCollection("users", database)
	errFindUsername := userCollection.FindOne(context.TODO(), bson.M{"username": username}).Err()

	if errFindUsername == nil {
		utils.JSONResponse(w, 409, nil, 0, "User exists")
		return
	}

	password, err := models.Hash(password)

	if err != nil {
		utils.JSONResponse(w, 500, nil, 0, "Register failed")
		return
	}
	filePath := "/static/" + fileName
	newUser := bson.M{"username": username, "password": password, "state": false, "last_login": nil, "img_path": filePath}
	_, err = userCollection.InsertOne(context.TODO(), newUser)
	if err != nil {
		utils.JSONResponse(w, 500, nil, 0, "Register failed")
		return
	}

	utils.JSONResponse(w, 200, nil, 1, "Register successfully")
}

func Login(c *gin.Context) {
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	if govalidator.IsNull(username) || govalidator.IsNull(password) {
		utils.JSONResponse(w, 400, nil, 0, "Bad request - data can not empty")
		return
	}
	var result bson.M
	userCollection := db.GetCollection("users", database)
	errFindUsername := userCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&result)

	if errFindUsername != nil {
		utils.JSONResponse(w, http.StatusOK, nil, 0, "User does not exists")
		return
	}
	hashedPassword := fmt.Sprintf("%v", result["password"])
	err := models.CheckPasswordHash(hashedPassword, password)

	if err != nil {
		utils.JSONResponse(w, http.StatusOK, nil, 0, "Password incorrect")
		return
	}
	token, err := CreateToken(username)
	if err != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, nil, 0, "Internal Server Error")
		return
	}

	filter := bson.M{"username": username}
	update := bson.M{"$set": bson.M{"state": true, "last_login": time.Now()}}
	_, updateErr := userCollection.UpdateOne(context.TODO(), filter, update)
	if updateErr != nil {
		log.Fatal("update login status failed")
	}

	utils.JSONResponse(w, http.StatusOK, token, 1, "success")
}

func GetHome(w http.ResponseWriter, r *http.Request) {
	content, err := ioutil.ReadFile("index.html")
	if err != nil {
		http.Error(w, "Could not open requested file", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "%s", content)
}

func UploadFile(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request

	//upload size
	err := r.ParseMultipartForm(200000) // grab the multipart form
	if err != nil {
		fmt.Fprintln(w, err)
	}

	//reading original file
	file, handler, err := r.FormFile("originalFile")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		return
	}
	defer file.Close()

	fileType := strings.Split(handler.Filename, ".")[1]
	fileName := uuid.NewString() + "." + fileType
	var fileMessage = FileMessage{
		Content:  handler.Filename,
		FilePath: "/static/" + fileName,
	}

	resFile, err := os.Create("./data/" + fileName)
	if err != nil {
		fmt.Fprintln(w, err)
		utils.JSONResponse(w, http.StatusInternalServerError, nil, 0, "Internal Server Error")
	}
	defer resFile.Close()
	if err == nil {
		io.Copy(resFile, file)
		defer resFile.Close()
		utils.JSONResponse(w, http.StatusOK, fileMessage, 1, "success")
	}
}

func DownloadFile(c *gin.Context) {
	filePath := c.Request.URL.Query().Get("file")
	fileName := c.Request.URL.Query().Get("name")

	c.Writer.Header().Set("Content-Disposition", "attachment; filename="+fileName)

	c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

	timeout := time.Duration(5) * time.Second
	transport := &http.Transport{
		ResponseHeaderTimeout: timeout,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, timeout)
		},
		DisableKeepAlives: true,
	}
	client := &http.Client{
		Transport: transport,
	}
	resp, err := client.Get("http://localhost:8080" + filePath)
	if err != nil {
		fmt.Println(err)
	}
	io.Copy(c.Writer, resp.Body)

	//downloadBytes, err := ioutil.ReadFile("./data/Capture.PNG")
	//if err != nil {
	//	fmt.Println(err)
	//}
	//mime := http.DetectContentType(downloadBytes)
	//
	//fileSize := len(string(downloadBytes))
	//
	//// Generate the server headers
	//c.Writer.Header().Set("Content-Type", mime)
	//c.Writer.Header().Set("Content-Disposition", "attachment; filename=" + "Capture.PNG" + "")
	//c.Writer.Header().Set("Expires", "0")
	//c.Writer.Header().Set("Content-Transfer-Encoding", "binary")
	//c.Writer.Header().Set("Content-Length", strconv.Itoa(fileSize))
	//c.Writer.Header().Set("Content-Control", "private, no-transform, no-store, must-revalidate")
	//http.ServeContent(c.Writer, c.Request, "Capture.PNG", time.Now(), bytes.NewReader(downloadBytes))
	//utils.JSONResponse(c.Writer, http.StatusOK, resp.Body, 1, "")
}
