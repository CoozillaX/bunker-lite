package main

import (
	"bunker-lite/routers"
	"fmt"
)

func main() {
	router := routers.InitRouter()
	router.Run(fmt.Sprintf(":%d", 8080))

	// http.HandleFunc("/api/new", std_api.New)
	// http.HandleFunc("/api/phoenix/login", std_api.Login)
	// http.HandleFunc("/api/phoenix/transfer_check_num", std_api.TransferCheckNum)
	// http.HandleFunc("/api/phoenix/transfer_start_type", std_api.TransferStartType)

	// fmt.Println("Server starts running...")
	// fmt.Println(http.ListenAndServe(":2333", nil))
}
