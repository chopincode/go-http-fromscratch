package httpd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

type conn struct {
	server *Server
	rwc    net.Conn
	lr     *io.LimitedReader
	bufr   *bufio.Reader
	bufw   *bufio.Writer // cached writer
}

func newConn(rwc net.Conn, server *Server) *conn {
	lr := &io.LimitedReader{R: rwc, N: 1 << 20}
	return &conn{
		server: server,
		rwc:    rwc,
		bufw:   bufio.NewWriterSize(rwc, 4<<10), // cache size 4kb
		lr:     lr,
		bufr:   bufio.NewReaderSize(lr, 4<<10),
	}
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
		if err = c.bufw.Flush(); err != nil {
			return
		}
	}
}

// empty implementation for now and will be implemented in later sections
func (c *conn) readRequest() (*Request, error) { return readRequest(c) }
func (c *conn) setupResponse() *response       { return setupResponse(c) }
func (c *conn) close()                         { c.rwc.Close() }
func handleErr(err error, c *conn)             { fmt.Println(err) }
