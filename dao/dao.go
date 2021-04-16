package dao

import (
	"database/sql"
	"v2/db"
)

var conn *sql.DB

func Init()  {
	conn = db.CreateCon()
}

type User struct {
	ID int `json:"id"`
	Name string `json:"name"`
}

func InsertUser(name string) sql.Result {
	var query = "insert into user(name) values (?)"
	r, err := conn.Exec(query, name)
	if err != nil{
		panic(err.Error())
	}
	return r
}

func GetCurrentUser() User  {
	var user User
	query := "select * from user order by id desc limit 1"
	err := conn.QueryRow(query).Scan(&user.ID, &user.Name)
	if err != nil{
		panic(err.Error())
	}
	return user
}

func InsertMessage(user User, message string) sql.Result {
	var query = "insert into message(user_id, msg_content) values (?, ?)"
	r, err := conn.Exec(query, user.ID, message)
	if err != nil{
		panic(err.Error())
	}
	return r
}
