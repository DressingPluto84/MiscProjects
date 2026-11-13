package main

import (
	context "context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var storage = make(map[int]string)
var counter int

func read(w http.ResponseWriter, r *http.Request, client AuthClient) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "GET":
		token := r.Header.Get("Authentication")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Missing Auth Token"}`))
			return
		}

		user := r.Header.Get("user")
		if user == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Missing User in Header"}`))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := client.VerifyToken(ctx, &Token{Token: token, User: user, Role: "user"})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Invalid Auth"}`))
			return
		}

		str := strings.TrimPrefix(r.URL.Path, "/read/")
		resource, err := strconv.Atoi(str)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Resource Not Found"}`))
			return
		}

		if resource < 0 {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Resource Not Found"}`))
			return
		}

		value := storage[resource]
		if value == "" {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error": "Rseource Not Found"}`))
			return
		}

		ret := fmt.Sprintf(`{"value": %s}`, value)
		w.WriteHeader(http.StatusAccepted)
		w.Write([]byte(ret))
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Only GET allowed"}`))
	}

}

func write(w http.ResponseWriter, r *http.Request, client AuthClient) {
	w.Header().Set("Content-Type", "application/json")

	switch r.Method {
	case "POST":
		token := r.Header.Get("Authentication")
		user := r.Header.Get("User")
		if token == "" {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Missing Auth Token"}`))
			return
		}

		if user == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Bad Request"}`))
			return
		}

		bts, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Bad Request"}`))
			return
		}

		val := make(map[string]string)
		err = json.Unmarshal(bts, &val)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": "Bad Request"}`))
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		payload := &Token{Token: token, User: user, Role: "user"}
		_, err = client.VerifyToken(ctx, payload)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error": "Unauthorized"}`))
			return
		}

		storage[counter] = fmt.Sprintf("%+v", val)
		counter += 1

		w.WriteHeader(http.StatusAccepted)
	default:
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not Found"}`))
	}
}

func main() {
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	client := NewAuthClient(conn)

	http.HandleFunc("/write", func(w http.ResponseWriter, r *http.Request) {
		write(w, r, client)
	})

	http.HandleFunc("/read/{object}", func(w http.ResponseWriter, r *http.Request) {
		read(w, r, client)
	})
	http.ListenAndServe(":3001", nil)
}
