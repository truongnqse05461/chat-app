package routes

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/novalagung/gubrak/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"v2/db"
	"v2/models"
	"v2/utils"
)

type M map[string]interface{}

const MESSAGE_NEW_USER = "New User"
const MESSAGE_CHAT = "Chat"
const MESSAGE_LEAVE = "Leave"

var connections = make([]*WebSocketConnection, 0)

type SocketPayload struct {
	Message string
	Type    string
	Room    string
	Path    string
}

type SocketResponse struct {
	From        string
	Type        string
	Message     string
	Avatar      string
	MessageType string
	FilePath    string
}

type WebSocketConnection struct {
	*websocket.Conn
	RoomName string
	Username string
}

func Contains(connections []*WebSocketConnection, roomName string, username string) (result bool) {
	result = false
	for _, product := range connections {
		if product.Username == username && product.RoomName == roomName {
			result = true
			break
		}
	}
	return result
}

type Author struct {
	Username string
	Avatar   string
}

type Message struct {
	Author      Author
	Content     string
	CreatedTime int64
	Type        string
	FilePath    string
}

type Conversation struct {
	Name        string
	Password    string
	Members     []string
	Messages    []Message
	RoomAvatar  string
	CreatedTime int64
}

func StartWebSocket(c *gin.Context) {
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	username := r.URL.Query().Get("username")
	roomName := r.URL.Query().Get("roomname")
	if !Contains(connections, roomName, username) {
		currentGorillaConn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
		if err != nil {
			http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
		}

		currentConn := WebSocketConnection{Conn: currentGorillaConn, RoomName: roomName, Username: username}
		connections = append(connections, &currentConn)
		//
		//conversationColl := db.GetCollection("conversations", database)
		//var result bson.M
		//errFindRoom := conversationColl.FindOne(context.TODO(), bson.M{"name": roomName}).Decode(&result)
		//if errFindRoom == nil{
		//	existUser := false
		//	switch reflect.TypeOf(result["members"]).Kind() {
		//	case reflect.Slice:
		//		s := reflect.ValueOf(result["members"])
		//
		//		for i := 0; i < s.Len(); i++ {
		//			u := fmt.Sprintf("%v", s.Index(i))
		//			if u == username {
		//				existUser = true
		//				break
		//			}
		//		}
		//	}
		//	if !existUser {
		//		filter := bson.M{"name": roomName}
		//		update := bson.M{"$push" : bson.M{"members":username}}
		//		_, err = conversationColl.UpdateOne(context.TODO(), filter, update)
		//
		//		if err != nil {
		//			utils.JSONResponse(w, 500, nil, 0,"Failed to start conversation")
		//		}
		//	}
		//}

		go handleIO(&currentConn, connections)
	}

}

func handleIO(currentConn *WebSocketConnection, connections []*WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	broadcastMessage(currentConn, MESSAGE_NEW_USER, "", "", "", "", "")

	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				broadcastMessage(currentConn, MESSAGE_LEAVE, "", "", "", "", "")
				ejectConnection(currentConn)
				return
			}

			log.Println("ERROR", err.Error())
			continue
		}
		var result bson.M
		userCollection := db.GetCollection("users", database)
		userCollection.FindOne(context.TODO(), bson.M{"username": currentConn.Username}).Decode(&result)

		avaPath := fmt.Sprintf("%v", result["img_path"])

		author := bson.D{
			{"username", currentConn.Username},
			{"avatar", avaPath},
		}
		message := bson.D{
			{"author", author},
			{"content", payload.Message},
			{"filePath", payload.Path},
			{"createdTime", time.Now().Unix()},
			{"type", payload.Type},
		}
		conversationColl := db.GetCollection("conversations", database)
		filter := bson.D{{"name", currentConn.RoomName}}
		update := bson.D{{"$push", bson.D{{"messages", message}}}}
		_, err = conversationColl.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			log.Fatalln(err.Error())
		}
		broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message, avaPath, payload.Type, payload.Room, payload.Path)
	}
}

func ejectConnection(currentConn *WebSocketConnection) {
	filtered := gubrak.From(connections).Reject(func(each *WebSocketConnection) bool {
		return each == currentConn
	}).Result()
	connections = filtered.([]*WebSocketConnection)
}

func broadcastMessage(currentConn *WebSocketConnection, kind, message string, avaPath string, msgType string, roomName string, filePath string) {
	for _, eachConn := range connections {
		if eachConn == currentConn || eachConn.RoomName != roomName {
			continue
		}

		eachConn.WriteJSON(SocketResponse{
			From:        currentConn.Username,
			Type:        kind,
			Message:     message,
			Avatar:      avaPath,
			MessageType: msgType,
			FilePath:    filePath,
		})
	}
}
func Clear(c *gin.Context) {
	var w http.ResponseWriter = c.Writer
	//var r *http.Request = c.Request

	conversationColl := db.GetCollection("conversations", database)
	_, err := conversationColl.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		log.Fatalln(err.Error())
	}
	utils.JSONResponse(w, http.StatusOK, nil, 0, "clear conversations document")
}

func GetMessages(c *gin.Context) {

	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	roomName := r.URL.Query().Get("roomname")
	conversationColl := db.GetCollection("conversations", database)
	var result Conversation
	findOptions := options.FindOne()
	// Sort by `price` field descending
	findOptions.SetSort(bson.D{{"createdTime", -1}})
	err := conversationColl.FindOne(context.TODO(), bson.D{{"name", roomName}}, findOptions).Decode(&result)

	if err != nil {
		log.Fatalln(err.Error())
	}
	//
	//for result.Next(context.TODO()) {
	//	var elem Conversation
	//	err := result.Decode(&elem)
	//	if err != nil {
	//		log.Fatalln(err.Error())
	//	}
	//	results = append(results, elem)
	//}
	//result.Close(context.TODO())

	utils.JSONResponse(w, http.StatusOK, result, 1, "success")
}

func InitChat(c *gin.Context) {
	var w http.ResponseWriter = c.Writer
	//var r *http.Request = c.Request
	//roomName := r.URL.Query().Get("roomname")
	conversationColl := db.GetCollection("conversations", database)
	var results []Conversation
	findOptions := options.Find()
	// Sort by `price` field descending
	findOptions.SetSort(bson.D{{"createdTime", -1}})
	result, err := conversationColl.Find(context.TODO(), bson.D{}, findOptions)

	if err != nil {
		log.Fatalln(err.Error())
	}

	for result.Next(context.TODO()) {
		var elem Conversation
		err := result.Decode(&elem)
		if err != nil {
			log.Fatalln(err.Error())
		}
		results = append(results, elem)
	}
	result.Close(context.TODO())

	utils.JSONResponse(w, http.StatusOK, results, 1, "success")
}

func CreateConversation(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	//roomName := r.URL.Query().Get("roomname")
	roomName := r.PostFormValue("roomname")
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	//upload size
	r.ParseMultipartForm(200000)

	//reading original file
	file, handler, errReadFile := r.FormFile("originalFile")
	fileName := ""
	if errReadFile != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(errReadFile.Error())
	} else {
		defer file.Close()

		fileType := strings.Split(handler.Filename, ".")[1]
		fileName = uuid.NewString() + "." + fileType
		resFile, errCreateFile := os.Create("./data/" + fileName)
		if errCreateFile != nil {
			fmt.Fprintln(w, errCreateFile)
		}
		defer resFile.Close()
		if errCreateFile == nil {
			io.Copy(resFile, file)
			defer resFile.Close()
		}
	}

	conversationColl := db.GetCollection("conversations", database)

	var result bson.M
	errFindRoom := conversationColl.FindOne(context.TODO(), bson.M{"name": roomName}).Decode(&result)
	if errFindRoom == nil {
		utils.JSONResponse(w, http.StatusFound, nil, 0, "Room name ("+roomName+") is exist")
	} else {
		if password != "" {
			password, _ = models.Hash(password)
		}
		filePath := ""
		if fileName != "" {
			filePath = "/static/" + fileName
		}
		newConversation := bson.M{
			"members":     bson.A{username},
			"name":        roomName,
			"password":    password,
			"createdTime": time.Now().Unix(),
			"roomAvatar":  filePath,
		}
		_, err := conversationColl.InsertOne(context.TODO(), newConversation)
		if err != nil {
			log.Fatalln(err.Error())
			utils.JSONResponse(w, http.StatusInternalServerError, nil, 0, "Internal Server Error")
		}
	}

	utils.JSONResponse(w, http.StatusOK, nil, 1, "success")
}

func JoinChat(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	//roomName := r.URL.Query().Get("roomname")
	username := r.PostFormValue("username")
	roomName := r.PostFormValue("roomname")
	password := r.PostFormValue("password")

	conversationColl := db.GetCollection("conversations", database)

	var result Conversation
	errFindRoom := conversationColl.FindOne(context.TODO(), bson.M{"name": roomName}).Decode(&result)
	if errFindRoom != nil {
		utils.JSONResponse(w, http.StatusInternalServerError, nil, 0, "Room name ("+roomName+") does not exist")
	} else {

		if result.Password != "" {
			err := models.CheckPasswordHash(result.Password, password)

			if err != nil {
				utils.JSONResponse(w, http.StatusOK, nil, 0, "Password incorrect")
				return
			}
		}
		filter := bson.M{"name": roomName}
		update := bson.M{"$push": bson.M{"members": username}}
		_, err := conversationColl.UpdateOne(context.TODO(), filter, update)

		if err != nil {
			utils.JSONResponse(w, 500, nil, 0, "Failed to start conversation")
		}
		utils.JSONResponse(w, http.StatusOK, nil, 1, "success")
	}

}
