package main

import (
	"bunker-lite/service/routers"
	"fmt"
)

func main() {
	router := routers.InitRouter()
	router.Run(fmt.Sprintf(":%d", 80))
}
