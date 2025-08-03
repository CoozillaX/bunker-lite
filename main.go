package main

import (
	"bunker-lite/routers"
	"fmt"
)

func main() {
	router := routers.InitRouter()
	router.Run(fmt.Sprintf(":%d", 8080))
}
