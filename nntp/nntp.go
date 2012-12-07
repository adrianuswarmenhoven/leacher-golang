package nntp

import (
	"bufio"
	"code.google.com/p/go-charset/charset"
	_ "code.google.com/p/go-charset/data"
	"errors"
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
)

var (
	responseRegexp  = regexp.MustCompile("(\\d{3}) (.*)")
	dotLineRegexp   = regexp.MustCompile("^\\.$")
	blankLineRegexp = regexp.MustCompile("^ *$")
)

type NntpConnection struct {
	net.Conn
	r *bufio.Reader
}

type NntpResponse struct {
	Code int
	Body string
}

type NntpGroupResponse struct {
	*NntpResponse
	Number int
	Low    int
	High   int
	Group  string
}

type NntpArticleResponse struct {
	*NntpResponse
	Headers map[string]string
	Bytes   [][]byte
}

type NntpError struct {
	Code int
	Body string
}

func (e *NntpError) Error() string {
	return fmt.Sprintf("failure response from nntp server code: %d, body: %s",
		e.Code, e.Body)
}

func responseFrom(s string) (*NntpResponse, error) {
	if matches := responseRegexp.FindStringSubmatch(s); matches != nil {
		code, err := strconv.Atoi(matches[1])
		if err != nil {
			return nil, err
		}
		body := matches[2]
		if code >= 400 {
			return nil, &NntpError{Code: code, Body: body}
		} else {
			log.Printf("got code: %d, body: %s", code, body)
			return &NntpResponse{Code: code, Body: body}, nil
		}
	}
	return nil, errors.New("could not parse response body: '" + s + "'")
}

func (conn *NntpConnection) send(a ...interface{}) {
	log.Printf("sending: %v", a)
	fmt.Fprintln(conn, a...)
}

func (conn *NntpConnection) response() (*NntpResponse, error) {
	resp, isPrefix, err := conn.r.ReadLine()
	if err != nil {
		return nil, err
	}
	if isPrefix {
		return nil, errors.New("unexpectedly long line!")
	}
	return responseFrom(string(resp))
}

func (conn *NntpConnection) multiLineResponse(re *regexp.Regexp) ([][]byte, error) {
	result := [][]byte{}
	for {
		bytes, isPrefix, err := conn.r.ReadLine()
		if err != nil {
			return nil, err
		}
		if isPrefix {
			return nil, errors.New("unexpectedly long line!")
		}
		if re.FindIndex(bytes) != nil {
			break
		}
		result = append(result, bytes)
	}
	log.Printf("got %d lines of response", len(result))
	return result, nil
}

func Connect(host string, port int) (*NntpConnection, error) {
	log.Printf("connecting to %s:%d", host, port)
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}
	r, err := charset.NewReader("ISO-8859-1", conn)
	if err != nil {
		return nil, err
	}
	result := &NntpConnection{conn, bufio.NewReader(r)}
	// get and ignore the welcome response
	if _, err := result.response(); err != nil {
		return nil, err
	}
	return result, nil
}

func (conn *NntpConnection) Authenticate(user, pass string) (*NntpResponse, error) {
	log.Printf("authenticating %s", user)
	conn.send("AUTHINFO USER", user)
	resp, err := conn.response()
	if err != nil {
		return nil, err
	}
	if resp.Code >= 300 && resp.Code <= 399 {
		conn.send("AUTHINFO PASS", pass)
		return conn.response()
	}
	return nil, &NntpError{Code: resp.Code, Body: resp.Body}
}

func (conn *NntpConnection) Group(group string) (*NntpGroupResponse, error) {
	log.Printf("changing to group %s", group)
	conn.send("GROUP", group)
	resp, err := conn.response()
	if err != nil {
		return nil, err
	}
	vals := strings.Split(resp.Body, " ")
	if len(vals) != 4 {
		return nil, errors.New("could not parse group response: " + resp.Body)
	}
	number, _ := strconv.Atoi(vals[0])
	low, _ := strconv.Atoi(vals[1])
	high, _ := strconv.Atoi(vals[2])
	return &NntpGroupResponse{resp, number, low, high, group}, nil
}

func parseHeaders(lines [][]byte) map[string]string {
	result := map[string]string{}
	for _, line := range lines {
		kvp := strings.Split(string(line), ": ")
		result[kvp[0]] = kvp[1]
	}
	log.Printf("got %d headers", len(result))
	return result
}

func (conn *NntpConnection) Article(id string) (*NntpArticleResponse, error) {
	log.Printf("getting article %s", id)
	conn.send("ARTICLE", id)
	resp, err := conn.response()
	if err != nil {
		return nil, err
	}
	headerLines, err := conn.multiLineResponse(blankLineRegexp)
	if err != nil {
		return nil, err
	}
	headers := parseHeaders(headerLines)
	bytes, err := conn.multiLineResponse(dotLineRegexp)
	if err != nil {
		return nil, err
	}
	return &NntpArticleResponse{resp, headers, bytes}, nil
}
