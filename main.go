package main

import (
	"bunker-core/protocol/mpay/android"
	"bunker-lite/service/routers"
	"fmt"
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

	router := routers.InitRouter()
	router.Run(fmt.Sprintf(":%d", 80))
}
