package http2

import (
	"fmt"

	"github.com/jimmyseraph/sparkle/easy_http"
)

func SendHttp2Request() {
	handler := easy_http.NewGet("https://127.0.0.1/hello")
	handler.SkipTLSCheck(true)
	handler.EnableHttp2(true)
	resp, err := handler.Execute()
	if err != nil {
		panic(err)
	}
	fmt.Printf("headers: %v\nproto: %s\nbody: %s\n", resp.Headers, resp.Proto, resp.Body)
}
