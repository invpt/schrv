package server

func HttpVersionNotSupported() HttpResponse {
	return HttpResponse{Status: 505, ResponseBody: "This server only supports HTTP/1.x."}
}

func NotImplemented() HttpResponse {
	return HttpResponse{Status: 501}
}

func InternalServerError() HttpResponse {
	return HttpResponse{Status: 500}
}

func NotFound() HttpResponse {
	return HttpResponse{Status: 404}
}

func BadRequest() HttpResponse {
	return HttpResponse{Status: 400}
}

func ExpectationFailed() HttpResponse {
	return HttpResponse{Status: 417}
}

func Ok(body string) HttpResponse {
	return HttpResponse{Status: 200, ResponseBody: body}
}
