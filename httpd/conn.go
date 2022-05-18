package httpd

import (
	"fmt"
	"log"
	"net"
)

type conn struct {
	server *Server
	rwc    net.Conn
}

func newConn(rwc net.Conn, server *Server) *conn {
	return &conn{server: server, rwc: rwc}
}

func (c *conn) serve() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("panic recoverred,err:%v\n", err)
		}
		c.close()
	}()

	// http1.1 support keep-alive long connection, so each connection may have multiple request
	// therefore a for loop is used to fetch all of them
	for {
		req, err := c.readRequest()
		if err != nil {
			handleErr(err, c)
			return
		}
		res := c.setupResponse()

		c.server.Handler.ServeHTTP(res, req)
	}
}

// empty implementation for now and will be implemented in later sections
func (c *conn) readRequest() (*Request, error) { return readRequest(c) }
func (c *conn) setupResponse() *response       { return nil }
func (c *conn) close()                         { c.rwc.Close() }
func handleErr(err error, c *conn)             { fmt.Println(err) }
