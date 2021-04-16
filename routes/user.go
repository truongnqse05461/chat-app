package routes

import (
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"net/http"
	"time"
	"v2/db"
	"v2/utils"
)

type User struct {
	Username string
	State bool
	LastLogin time.Time
}

func GetUserInfo(c *gin.Context)  {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	//claims, err := ValidToken(r)
	//if err != nil {
	//	utils.JSONResponse(w, http.StatusUnauthorized, "Unauthorized")
	//	return
	//}
	username := r.URL.Query().Get("username")
	var result bson.M
	userCollection := db.GetCollection("users", database)
	updateErr := userCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&result)
	if updateErr != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, nil, 0,"Internal Server Error")
	}
	utils.JSONResponse(w, http.StatusOK, result, 1, "success")
}
func ClearUser(c *gin.Context)  {
	var w http.ResponseWriter = c.Writer
	//var r *http.Request = c.Request

	userCollection := db.GetCollection("users", database)
	_ , err := userCollection.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		log.Fatalln(err.Error())
	}
	utils.JSONResponse(w, http.StatusOK, nil, 1, "clear users document")
}

