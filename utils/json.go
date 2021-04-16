package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	Code int
	Message string
	Data interface{}
}
func JSONResponse(w http.ResponseWriter, statusCode int, data interface{}, code int, message string)  {
	resp := Response{
		Data: data,
		Code: code,
		Message: message,
	}
	w.WriteHeader(statusCode)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		fmt.Fprintf(w, "%s", err.Error())
	}
}