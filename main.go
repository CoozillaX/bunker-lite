package main

import (
	"bunker-lite/std_api"
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/api/new", std_api.New)
	http.HandleFunc("/api/phoenix/login", std_api.Login)
	http.HandleFunc("/api/phoenix/transfer_check_num", std_api.TransferCheckNum)
	http.HandleFunc("/api/phoenix/transfer_start_type", std_api.TransferStartType)

	fmt.Println("Server starts running...")
	fmt.Println(http.ListenAndServe(":2333", nil))
}
