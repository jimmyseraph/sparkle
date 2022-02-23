package performance

import (
	"fmt"
	"os"
	"time"

	"github.com/jimmyseraph/sparkle/easy_http"
)

func LoadTest(vuser int, seconds int) {
	var body = `{}`
	result := make(chan string, vuser)
	var stop bool = false
	time.AfterFunc(time.Duration(seconds)*time.Second, func() {
		stop = true
	})
	for i := 0; i < vuser; i++ {
		go func() {
			for !stop {
				handler := easy_http.NewPost("https://xxx", body)
				handler.Headers["Content-Type"] = []string{"application/json"}
				handler.Headers["apikey"] = []string{"123"}
				handler.Headers["x-transaction-id"] = []string{""}
				resp, err := handler.Execute()
				if err != nil {
					result <- fmt.Sprintf("%v, %s", resp.Duration, err.Error())
				} else {
					result <- fmt.Sprintf("%v, %s", resp.Duration, resp.Body)
				}

			}
		}()
	}

	f, _ := os.Create("resp.csv")
	defer f.Close()
	for !stop {
		f.WriteString(<-result)
	}
}
