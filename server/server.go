package server

import (
	"log"
	"net"
	"strconv"
	"strings"
)

type HttpResponse struct {
	Status       uint16
	Headers      map[string][]string
	ResponseBody string
}

type HttpRequest struct {
	RequestLine RequestLine
	Headers     map[string][]string
	RequestBody string
}

type HeaderField struct {
	Name  string
	Value string
}

type RequestLine struct {
	Method  string
	Target  []string
	Version uint8
}

type StatusLine struct {
	Version uint8
	Status  uint16
}

func SimpleResponse(code uint16, body string) HttpResponse {
	return HttpResponse{Status: code, Headers: map[string][]string{}, ResponseBody: body}
}

func EmptyResponse() HttpResponse {
	return HttpResponse{Status: 200, Headers: map[string][]string{}, ResponseBody: ""}
}

type RequestHandler func(request HttpRequest) (response HttpResponse)

type Server struct {
	listener net.Listener
	handlers []func(request HttpRequest) (bool, HttpResponse)
}

func Bind(address string) (Server, error) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return Server{}, err
	}

	return Server{listener: listener}, nil
}

func (server *Server) Get(path []string, handler RequestHandler) {
	server.handlers = append(server.handlers, func(request HttpRequest) (bool, HttpResponse) {
		if len(path) != len(request.RequestLine.Target) {
			return false, HttpResponse{}
		}

		for i := 0; i < len(path); i++ {
			if path[i] != request.RequestLine.Target[i] {
				return false, HttpResponse{}
			}
		}

		return true, handler(request)
	})
}

func (server *Server) CustomHandler(handler func(request HttpRequest) (bool, HttpResponse)) {
	server.handlers = append(server.handlers, handler)
}

func (server *Server) Serve() error {
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			return err
		}

		log.Print("Accepted connection from ", conn.RemoteAddr())

		go server.handleConnection(conn)
	}
}

func (server *Server) handleConnection(conn net.Conn) {
	chars := newCharStream(conn)

	request, err := parseHttpMessage(&chars)
	if err != nil {
		conn.Close()
		log.Print("Failed to parse HTTP message: ", err)
		return
	}

	log.Print("Parsed request.")
	log.Print("Handling request.")

	response, err := server.handleRequest(request, &chars)
	if err != nil {
		conn.Close()
		log.Print("Request handler threw error: ", err)
		return
	}

	log.Print("Response generated.")

	responseText := unparse(response)

	log.Print("Response serialized.")

	_, err = conn.Write([]byte(responseText))
	if err != nil {
		conn.Close()
		log.Print("Error occurred while writing response to connection: ", err)
		return
	}

	log.Print("Response sent.")

	err = conn.Close()
	if err != nil {
		log.Print("Error occurred while closing connection: ", err)
		return
	}

	log.Print("Connection closed.")
}

func (server *Server) handleRequest(request HttpRequest, chars *charStream) (HttpResponse, error) {
	if request.RequestLine.Version>>4 != 1 {
		return HttpVersionNotSupported(), nil
	}

	if len(request.Headers["Host"]) != 1 {
		return BadRequest(), nil
	}

	if len(request.Headers["Content-Length"]) > 1 {
		return BadRequest(), nil
	} else if len(request.Headers["Content-Length"]) == 1 {
		contentLength, err := strconv.ParseInt(request.Headers["Content-Length"][0], 10, 0)
		if err != nil || contentLength < 0 {
			return BadRequest(), nil
		}

		request.RequestBody, err = chars.Read(uint(contentLength))
		if err != nil {
			return HttpResponse{}, err
		}
	}

	for _, expectation := range request.Headers["Expect"] {
		if expectation == "100-continue" {
			panic("Continuation is unimplemented")
		} else {
			return ExpectationFailed(), nil
		}
	}

	switch strings.ToUpper(request.RequestLine.Method) {
	case "GET":
		for _, handler := range server.handlers {
			handled, response := handler(request)

			if handled {
				return response, nil
			}
		}

		return NotFound(), nil
	case "HEAD":
		// is this too hacky? should users be able to know whether a request is GET or HEAD?

		for _, handler := range server.handlers {
			handled, response := handler(request)

			if handled {
				response.ResponseBody = ""
				return response, nil
			}
		}

		return NotFound(), nil
	default:
		return NotImplemented(), nil
	}
}
