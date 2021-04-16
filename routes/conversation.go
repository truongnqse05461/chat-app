package routes

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/novalagung/gubrak/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
	"v2/db"
	"v2/utils"
)


type M map[string]interface{}

const MESSAGE_NEW_USER = "New User"
const MESSAGE_CHAT = "Chat"
const MESSAGE_LEAVE = "Leave"

var connections = make([]*WebSocketConnection, 0)


type SocketPayload struct {
	Message string
	Type string
}

type SocketResponse struct {
	From    string
	Type    string
	Message string
	Avatar string
	MessageType string
}

type WebSocketConnection struct {
	*websocket.Conn
	RoomName string
	Username string
}

type Author struct {
	Username string
	Avatar string
}

type Message struct {
	Author Author
	Content string
	CreatedTime int64
	Type string
}

type Conversation struct {
	Name string
	Members []string
	Messages []Message
}

func StartWebSocket(c *gin.Context)  {
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	currentGorillaConn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}
	username := r.URL.Query().Get("username")
	roomName := r.URL.Query().Get("roomname")
	currentConn := WebSocketConnection{Conn: currentGorillaConn, RoomName: roomName, Username: username}
	connections = append(connections, &currentConn)

	conversationColl := db.GetCollection("conversations", database)
	var result bson.M
	errFindRoom := conversationColl.FindOne(context.TODO(), bson.M{"name": roomName}).Decode(&result)
	if errFindRoom == nil{
		existUser := false
		switch reflect.TypeOf(result["members"]).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(result["members"])

			for i := 0; i < s.Len(); i++ {
				u := fmt.Sprintf("%v", s.Index(i))
				if u == username {
					existUser = true
					break
				}
			}
		}
		if !existUser {
			filter := bson.M{"name": roomName}
			update := bson.M{"$push" : bson.M{"members":username}}
			_, err = conversationColl.UpdateOne(context.TODO(), filter, update)

			if err != nil {
				utils.JSONResponse(w, 500, nil, 0,"Failed to start conversation")
			}
		}
	}else {
		newConversation := bson.M{"members":bson.A{username}, "name" : roomName}
		_, err = conversationColl.InsertOne(context.TODO(), newConversation)

	}

	go handleIO(&currentConn, connections)
}

func handleIO(currentConn *WebSocketConnection, connections []*WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("ERROR", fmt.Sprintf("%v", r))
		}
	}()

	broadcastMessage(currentConn, MESSAGE_NEW_USER, "", "", "")

	for {
		payload := SocketPayload{}
		err := currentConn.ReadJSON(&payload)
		if err != nil {
			if strings.Contains(err.Error(), "websocket: close") {
				broadcastMessage(currentConn, MESSAGE_LEAVE, "", "", "")
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
		broadcastMessage(currentConn, MESSAGE_CHAT, payload.Message, avaPath, payload.Type)
	}
}

func ejectConnection(currentConn *WebSocketConnection) {
	filtered := gubrak.From(connections).Reject(func(each *WebSocketConnection) bool {
		return each == currentConn
	}).Result()
	connections = filtered.([]*WebSocketConnection)
}

func broadcastMessage(currentConn *WebSocketConnection, kind, message string, avaPath string, msgType string) {
	for _, eachConn := range connections {
		if eachConn == currentConn {
			continue
		}

		eachConn.WriteJSON(SocketResponse{
			From:    currentConn.Username,
			Type:    kind,
			Message: message,
			Avatar: avaPath,
			MessageType: msgType,
		})
	}
}
func Clear(c *gin.Context)  {
	var w http.ResponseWriter = c.Writer
	//var r *http.Request = c.Request

	conversationColl := db.GetCollection("conversations", database)
	_ , err := conversationColl.DeleteMany(context.TODO(), bson.M{})
	if err != nil {
		log.Fatalln(err.Error())
	}
	utils.JSONResponse(w, http.StatusOK, nil, 0,"clear conversations document")
}

func GetMessages(c *gin.Context)  {
	var w http.ResponseWriter = c.Writer
	var r *http.Request = c.Request
	roomName := r.URL.Query().Get("roomname")
	conversationColl := db.GetCollection("conversations", database)
	var result Conversation
	findOptions := options.FindOne()
	// Sort by `price` field descending
	findOptions.SetSort(bson.D{{"createdTime", -1}})
	err := conversationColl.FindOne(context.TODO(), bson.D{{"name",roomName}}, findOptions).Decode(&result)

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

	utils.JSONResponse(w, http.StatusOK,result.Messages, 1,"success" )
}