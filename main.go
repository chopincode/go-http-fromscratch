package main

import (
	"bytes"
	"fmt"
	"go-http-from-scratch/httpd"
	"io"
)

type myHandler struct{}

func (*myHandler) ServeHTTP(w httpd.ResponseWriter, r *httpd.Request) {
	// store user header information in to buff
	buff := &bytes.Buffer{}

	// test request
	fmt.Fprintf(buff, "[query]name=%s\n", r.Query("name"))
	fmt.Fprintf(buff, "[query]token=%s\n", r.Query("token"))
	fmt.Fprintf(buff, "[cookie]foo1=%s\n", r.Cookie("foo1"))
	fmt.Fprintf(buff, "[cookie]foo2=%s\n", r.Cookie("foo2"))
	fmt.Fprintf(buff, "[Header]User-Agent=%s\n", r.Header.Get("User-Agent"))
	fmt.Fprintf(buff, "[Header]Proto=%s\n", r.Proto)
	fmt.Fprintf(buff, "[Header]Method=%s\n", r.Method)
	fmt.Fprintf(buff, "[Addr]Addr=%s\n", r.RemoteAddr)
	fmt.Fprintf(buff, "[Request]%+v\n", r)

	// manually sent response message back
	io.WriteString(w, "HTTP/1.1 200 OK\r\n")
	io.WriteString(w, fmt.Sprintf("Content-Length: %d\r\n", buff.Len()))
	io.WriteString(w, "\r\n")
	io.Copy(w, buff) //send buffered information back to client
}

func main() {
	server := httpd.Server{
		Addr:    "127.0.0.1:8080",
		Handler: new(myHandler),
	}
	panic(server.ListenAndServe())
	// test using
	// curl "127.0.0.1:8080?name=gu&token=123" -b "foo1=bar1;foo2=bar2;" -i
}
