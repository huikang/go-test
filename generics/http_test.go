package generics

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestHTTPGenerics(t *testing.T) {

	url := "http://localhost:59357/static-server-2/debug?env=dump"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	res, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(body))
}
