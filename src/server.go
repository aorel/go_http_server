package main

import (
	"fmt"
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

var fileTypes = map[string]string{
	".css":  "text/css",
	".gif":  "image/gif",
	".html": "text/html",
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".js":   "application/x-javascript",
	".pdf":  "application/pdf",
	".png":  "image/png",
	".svg":  "image/svg+xml",
	".swf":  "application/x-shockwave-flash",
	".xml":  "text/xml",
}

var statusMessage = map[int]string{
	200: "OK",
	400: "Bad Request",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
}

type Headers map[string]string

type RequestLine struct {
	Method  string
	URI     string
	Version string
}

type Request struct {
	StartingLine RequestLine
	Headers      Headers
}

func (req *Request) Parse(requestString string) {

	requestSlice := strings.Split(requestString, "\r\n")
	if len(requestSlice) < 2 {
		return
	}

	startingLine := requestSlice[0]
	startingLineSlice := strings.Split(startingLine, " ")

	req.StartingLine.Method = startingLineSlice[0]
	req.StartingLine.URI = startingLineSlice[1]
	req.StartingLine.Version = startingLineSlice[2]
}

type ResponseLine struct {
	Version string
	Status  int
	Message string
}

type Response struct {
	ResponseLine ResponseLine
	Headers      Headers
	Body         []byte
}

func (response *Response) Status(code int) {
	response.ResponseLine.Status = code
	response.ResponseLine.Message = statusMessage[code]
}

func (response *Response) AddHeader(key, value string) {
	response.Headers[key] = value
}

func (response *Response) AddBody(file []byte) {
	response.Body = file
}

func (response *Response) BuildHeaders() (str string) {
	str += response.ResponseLine.Version + " " +
		strconv.FormatInt(int64(response.ResponseLine.Status), 10) + " " +
		response.ResponseLine.Message + "\r\n"

	for key, value := range response.Headers {
		str += key + ": " + value + "\r\n"
	}

	str += "Date: " + time.Now().String() + "\r\n"
	str += "Server: go_http_server" + "\r\n"
	str += "Connection: close" + "\r\n"
	str += "\r\n"
	return
}

type Handler struct {
	Conn     net.Conn
	Request  Request
	Response Response
}

func (handler *Handler) Close() {
	handler.Conn.Write([]byte(handler.Response.BuildHeaders()))
	if len(handler.Response.Body) != 0 {
		handler.Conn.Write(handler.Response.Body)
	}
	handler.Conn.Close()
}

func (handler *Handler) Handle(root string) {
	defer handler.Close()
	for {
		buf := make([]byte, 1024)

		if _, err := handler.Conn.Read(buf); err != nil {
			// fmt.Println("Failed to read:", err.Error())
			break
		}

		handler.Request.Headers = make(map[string]string)
		handler.Response.Headers = make(map[string]string)

		handler.Request.Parse(string(buf))
		handler.Response.ResponseLine.Version = handler.Request.StartingLine.Version

		method := handler.Request.StartingLine.Method
		uri := handler.Request.StartingLine.URI

		if strings.Contains(uri, "../") {
			handler.Response.Status(400)
			break
		}

		isDirectory := strings.HasSuffix(uri, "/")
		if isDirectory {
			uri = uri + "index.html"
			_, err := os.Open(root + uri)
			if err != nil {
				handler.Response.Status(403)
				break
			}
		}

		paramsIndex := strings.LastIndex(uri, "?")
		if paramsIndex > -1 {
			uri = uri[:paramsIndex]
		}

		_uri, err := url.QueryUnescape(uri)
		if err != nil {
			handler.Response.Status(400)
			break
		}

		_path := root + _uri
		path, _ := filepath.Abs(_path)

		switch {
		case method == "GET":
			file, err := ioutil.ReadFile(path)
			if err != nil {
				handler.Response.Status(404)
				break
			}
			handler.Response.Status(200)

			fileExt := filepath.Ext(path)
			typeString, ok := fileTypes[fileExt]
			if ok {
				handler.Response.AddHeader("Content-Type", typeString)
			}
			handler.Response.AddHeader("Content-Length", strconv.Itoa(len(file)))
			handler.Response.AddBody(file)

		case method == "HEAD":
			file, err := os.Open(path)
			if err != nil {
				handler.Response.Status(200)
				break
			}
			handler.Response.Status(200)
			fi, err := file.Stat()
			handler.Response.AddHeader("Content-Length", strconv.FormatInt(fi.Size(), 10))
		default:
			handler.Response.Status(405)
			break
		}

		break
	}
}

func StartServer(port, root string) {
	fmt.Println("Server starting on port", port, "...")
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		fmt.Println("Failed to start:", err.Error())
		return
	}

	fmt.Println("Root", root)
	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("New connection failed:", err.Error())
			conn.Close()
			continue
		}
		// fmt.Println("New connection success:", conn.RemoteAddr())

		handler := new(Handler)
		handler.Conn = conn
		go handler.Handle(root)
	}
}
