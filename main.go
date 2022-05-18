package main

import (
	"fmt"
	"go-http-from-scratch/httpd"
)

type myHandler struct{}

func (*myHandler) ServeHTTP(w httpd.ResponseWriter, r *httpd.Request) {
	fmt.Println("hello world")
}

func main() {
	server := httpd.Server{
		Addr:    "127.0.0.1:8080",
		Handler: new(myHandler),
	}
	panic(server.ListenAndServe())
}
