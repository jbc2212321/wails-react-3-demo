package update

import (
	"fmt"
	"io"
	"net/http"
	"testing"
)

func TestUpdate(t *testing.T) {
	//confirmAndSelfUpdate()
	DoSelfUpdate()
}

func TestUpdate2(t *testing.T) {
	update()
}

func TestDemo(t *testing.T) {
	url := "https://api.github.com/repos/jbc2212321/wails-react-3-demo/releases"
	method := "GET"

	client := &http.Client{}
	req, err := http.NewRequest(method, url, nil)

	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Authorization", "")

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

func TestUp(t *testing.T) {
	AppUpdate()
}
