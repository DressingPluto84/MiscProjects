package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

var numRoutines = 8
var lock sync.Mutex
var idx = 0

func getRequest(w http.ResponseWriter, r *http.Request, arr [](chan map[string]string)) {
	lock.Lock()
	ind := idx % numRoutines
	idx += 1
	lock.Unlock()

	w.Header().Add("Content-Type", "application/json")
	bd, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	m := make(map[string]string)
	err = json.Unmarshal(bd, &m)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Fprintf(os.Stdout, "Sending Body to: %d\n", ind)
	arr[ind] <- m
	ret := <-arr[ind]
	if ret["resp"] == "Fail" {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		w.WriteHeader(http.StatusAccepted)
	}
}

func handleRequest(c chan map[string]string, i int) {
	for j := 0; j < 1; j += 0 {
		m := <-c
		cnt := 0
		for k, v := range m {
			fmt.Fprintf(os.Stdout, "%d-> %s: %s\n", i, k, v)
			cnt += 1
		}

		resp := make(map[string]string)
		if cnt < 2 {
			resp["resp"] = "Fail"
		} else {
			resp["resp"] = "Pass"
		}

		c <- resp
	}
}

func main() {
	var arr [](chan map[string]string)

	for i := 0; i < numRoutines; i++ {
		c := make(chan map[string]string)
		arr = append(arr, c)
		go handleRequest(c, i)
	}

	http.HandleFunc("/app", func(w http.ResponseWriter, r *http.Request) {
		getRequest(w, r, arr)
	})
	http.ListenAndServe(":8080", nil)
}
