package main

import (
	"fmt"
	"net/http"
	"time"
)

func main() {
	start := time.Now()
	resp, err := http.Get("http://localhost:8080/bootstrap") // adjust url if needed, actually we can just bench the functions
	fmt.Println(time.Since(start), err, resp)
}
