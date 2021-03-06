package httpd

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
)

type Request struct {
	Method      string            // http methods, eg, GET, POST, PUT
	URL         *url.URL          // URL
	Proto       string            // Protocol and version
	Header      Header            // http header
	Body        io.Reader         // to read http message body
	RemoteAddr  string            // client address
	RequestURI  string            // url in string format
	conn        *conn             // connection that create this request
	cookies     map[string]string // cookie storage
	queryString map[string]string // query string storage
}

func readRequest(c *conn) (r *Request, err error) {
	r = new(Request)
	r.conn = c
	r.RemoteAddr = c.rwc.RemoteAddr().String()
	// read the first line, eg. GET / index?name=hello HTTP/1.1
	line, err := readLine(c.bufr)
	if err != nil {
		return
	}

	_, err = fmt.Sscanf(string(line), "%s%s%s", &r.Method, &r.RequestURI, &r.Proto)
	if err != nil {
		return
	}

	// convert string url to url.URL
	r.URL, err = url.ParseRequestURI(r.RequestURI)
	if err != nil {
		return
	}

	// parse query string
	r.parseQuery()
	// read header
	r.Header, err = readHeader(c.bufr)
	if err != nil {
		return
	}

	const noLimit = (1 << 63) - 1
	r.conn.lr.N = noLimit

	r.setupBody()
	return r, nil
}

func readLine(bufr *bufio.Reader) ([]byte, error) {
	p, isPrefix, err := bufr.ReadLine()
	if err != nil {
		return p, err
	}
	var l []byte
	for isPrefix {
		l, isPrefix, err = bufr.ReadLine()
		if err != nil {
			break
		}
		p = append(p, l...)
	}

	return p, err
}

func (r *Request) parseQuery() {
	// r.URL.RawQuery = "name=hello&token=1234"
	r.queryString = parseQuery(r.URL.RawQuery)
}

func parseQuery(RawQuery string) map[string]string {
	parts := strings.Split(RawQuery, "&")
	queries := make(map[string]string, len(parts))
	for _, part := range parts {
		index := strings.IndexByte(part, '=')
		if index == -1 || index == len(part)-1 {
			continue
		}
		queries[strings.TrimSpace(part[:index])] = strings.TrimSpace(part[index+1:])
	}
	return queries
}

func readHeader(bufr *bufio.Reader) (Header, error) {
	header := make(Header)
	for {
		line, err := readLine(bufr)
		if err != nil {
			return nil, err
		}
		//if we reach /r/n/r/n???it means we read the eof of the header message
		if len(line) == 0 {
			break
		}
		//example: Connection: keep-alive
		i := bytes.IndexByte(line, ':')
		if i == -1 {
			return nil, errors.New("unsupported protocol")
		}
		if i == len(line)-1 {
			continue
		}
		k, v := string(line[:i]), strings.TrimSpace(string(line[i+1:]))
		header[k] = append(header[k], v)
	}
	return header, nil
}

type eofReader struct{}

// implement io.Reader interface
func (er *eofReader) Read([]byte) (n int, err error) {
	return 0, io.EOF
}

func (r *Request) setupBody() {
	r.Body = new(eofReader)
}

// Query search for query string
func (r *Request) Query(name string) string {
	return r.queryString[name]
}

// Cookie search for cookie
func (r *Request) Cookie(name string) string {
	if r.cookies == nil {
		r.parseCookies()
	}
	return r.cookies[name]
}

func (r *Request) parseCookies() {
	if r.cookies != nil {
		return
	}
	r.cookies = make(map[string]string)
	rawCookies, ok := r.Header["Cookie"]
	if !ok {
		return
	}
	for _, line := range rawCookies {
		//example(line): uuid=12314753; tid=1BDB9E9; HOME=1(see http request struct above)
		kvs := strings.Split(strings.TrimSpace(line), ";")
		if len(kvs) == 1 && kvs[0] == "" {
			continue
		}
		for i := 0; i < len(kvs); i++ {
			//example(kvs[i]): uuid=12314753
			index := strings.IndexByte(kvs[i], '=')
			if index == -1 {
				continue
			}
			r.cookies[strings.TrimSpace(kvs[i][:index])] = strings.TrimSpace(kvs[i][index+1:])
		}
	}
	return
}
