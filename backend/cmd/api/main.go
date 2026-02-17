package main

import (
	"log"

	apihttp "github.com/bajiaozhi/w-mma/backend/internal/http"
)

func main() {
	srv := apihttp.NewServer()
	if err := srv.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
