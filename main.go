package main

import (
	"bunker-core/protocol/mpay/android"
	"bunker-lite/api"
	"fmt"
	"net/http"
	"time"
)

func main() {
	// Resource downloads belong to startup/background work, never an HTTP login.
	go func() {
		retry := time.Minute
		for {
			next := 15 * time.Minute
			if err := android.RefreshPublishedPackages(); err != nil {
				fmt.Printf("Android package refresh failed: %v\n", err)
				next = retry
				retry = min(retry*2, 15*time.Minute)
			} else {
				retry = time.Minute
			}
			time.Sleep(next)
		}
	}()

	http.HandleFunc("/api/new", api.New)
	http.HandleFunc("/api/phoenix/login", api.Login)
	http.HandleFunc("/api/phoenix/transfer_check_num", api.TransferCheckNum)
	http.HandleFunc("/api/phoenix/transfer_start_type", api.TransferStartType)

	fmt.Println("Server starts running...")
	fmt.Println(http.ListenAndServe(":2333", nil))
}
