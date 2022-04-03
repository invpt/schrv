package main

import (
	"errors"
	"gggg/server"
	"log"
	"os"
	"strings"
)

func main() {
	sv, err := server.Bind(":8000")
	if err != nil {
		log.Fatal("Failed to create server: ", err)
	}

	sv.CustomHandler(func(request server.HttpRequest) (bool, server.HttpResponse) {
		path := ""
		for i := 0; i < len(request.RequestLine.Target); i++ {
			path += "/"
			if strings.Contains(request.RequestLine.Target[i], "/") {
				return true, server.InternalServerError()
			}
			path += request.RequestLine.Target[i]
		}

		data, err := os.ReadFile(path[1:])
		if err != nil {
			return true, server.NotFound()
		}

		respBody := string(data)

		return true, server.HttpResponse{Status: 200, ResponseBody: respBody}
	})

	log.Print("Starting server")
	err = sv.Serve()
	if err != nil {
		log.Fatal("Error encountered while running server: ", err)
	}
}
