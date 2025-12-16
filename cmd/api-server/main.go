package main

import (
	"fmt"
)

func main() {
	r := InitServer()

	fmt.Println("Server starting on :8080...")
	r.Run(":8081")
}
