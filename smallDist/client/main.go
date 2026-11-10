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

func main() {
	url := "http://localhost:3000/login"

	user := "Anish"
	pass := "Poo12345@"
	/*
		var opt int

		fmt.Print("Create Account (0), Login (1): ")
		fmt.Scanln(&opt)
		fmt.Print("Enter Username: ")
		fmt.Scanln(&user)
		fmt.Print("Enter Password: ")
		fmt.Scanln(&pass)
	*/

	p := loginPayload{user, pass}
	body, err := json.Marshal(p)
	if err != nil {
		fmt.Println("Error encoding user and password")
		return
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		fmt.Println(err)
		return
	}
	req.Header.Add("Content-Type", "application/json")

	resp, _ := client.Do(req)

	r, _ := io.ReadAll(resp.Body)
	respBody := make(map[string]string)
	_ = json.Unmarshal(r, &respBody)
	fmt.Println(respBody)
}
