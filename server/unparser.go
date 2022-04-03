package server

import (
	"bytes"
	"strconv"
)

var reasonPhrases = map[uint16]string{
	100: "Continue",
	101: "Switching Protocols",
	200: "OK",
	201: "Created",
	202: "Accepted",
	203: "Non-Authoritative Information",
	204: "No Content",
	205: "Reset Content",
	300: "Multiple Choices",
	301: "Moved Permanently",
	302: "Found",
	303: "See Other",
	305: "Use Proxy",
	307: "Temporary Redirect",
	400: "Bad Request",
	402: "Payment Required",
	403: "Forbidden",
	404: "Not Found",
	405: "Method Not Allowed",
	406: "Not Acceptable",
	408: "Request Timeout",
	409: "Conflict",
	410: "Gone",
	411: "Length Required",
	413: "Payload Too Large",
	414: "URI Too Long",
	415: "Unsupported Media Type",
	417: "Expectation Failed",
	426: "Upgrade Required",
	500: "Internal Server Error",
	501: "Not Implemented",
	502: "Bad Gateway",
	503: "Service Unavailable",
	504: "Gateway Timeout",
	505: "HTTP Version Not Supported",
}

func unparse(response HttpResponse) string {
	buffer := new(bytes.Buffer)

	buffer.WriteString("HTTP/1.1")

	buffer.WriteByte(' ')
	buffer.WriteString(strconv.FormatInt(int64(response.Status), 10))
	buffer.WriteByte(' ')
	buffer.WriteString(reasonPhrases[response.Status])
	buffer.WriteByte(' ')
	buffer.WriteString("\r\n")

	for headerName, headerValues := range response.Headers {
		for _, headerValue := range headerValues {
			buffer.WriteString(headerName)
			buffer.WriteByte(':')
			buffer.WriteString(headerValue)
			buffer.WriteString("\r\n")
		}
	}
	buffer.WriteString("\r\n")
	buffer.WriteString(response.ResponseBody)

	return buffer.String()
}
