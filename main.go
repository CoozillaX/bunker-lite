package main

import (
	"bunker-lite/api"
	"fmt"
)

func main() {
	router.Run(fmt.Sprintf(":%d", 8080))
	router := api.InitRouter()
}
