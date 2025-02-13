package main

import (
	"bunker-lite/api"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/api/new", api.New)
	http.HandleFunc("/api/phoenix/login", api.Login)
	http.HandleFunc("/api/phoenix/transfer_check_num", api.TransferCheckNum)
	http.HandleFunc("/api/phoenix/transfer_start_type", api.TransferStartType)

	fmt.Println("Server starts running...")
	http.ListenAndServe(":2333", nil)
}
