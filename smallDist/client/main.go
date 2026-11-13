package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type loginPayload struct {
	User string `json:"user"`
	Pass string `json:"pass"`
}

type writePayload struct {
	Thing string `json:"thing"`
}

func main() {
	auth := "http://localhost:3000/login"
	service := "http://localhost:3001/write"

	user := "user"
	pass := "pass"

	p := loginPayload{user, pass}
	body, err := json.Marshal(p)
	if err != nil {
		fmt.Println("Error encoding user and password")
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", auth, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, _ := client.Do(req)

	r, _ := io.ReadAll(resp.Body)
	respBody := make(map[string]string)
	_ = json.Unmarshal(r, &respBody)
	jwt := respBody["auth"]

	if jwt == "" {
		fmt.Println(respBody)
		fmt.Println("Failed")
	}

	pay := writePayload{"thingy"}
	body, err = json.Marshal(pay)

	req, err = http.NewRequest("POST", service, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authentication", jwt)
	req.Header.Add("User", "Janish")
	_, _ = client.Do(req)

	req, err = http.NewRequest("GET", "http://localhost:3001/read/0", bytes.NewBuffer([]byte("")))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authentication", jwt)
	req.Header.Add("User", "Janish")

	resp, err = client.Do(req)
	r, _ = io.ReadAll(resp.Body)
	respBody = make(map[string]string)
	_ = json.Unmarshal(r, &respBody)

	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(resp.StatusCode)
		fmt.Println(respBody)
	}
}
