package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

func sendReq(i int) {
	m := make(map[string]string)
	m["a"] = "1"

	if i%2 == 0 {
		m["b"] = "2"
	}
	p, err := json.Marshal(&m)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := http.Post("http://localhost:8080/app", "applicaiton/json", bytes.NewBuffer(p))
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(i, resp.Status)
}

func main() {
	var wg sync.WaitGroup

	for i := 0; i < 12; i++ {
		wg.Add(1)
		wg.Go(func() {
			defer wg.Done()
			sendReq(i)
		})
	}

	wg.Wait()
}
