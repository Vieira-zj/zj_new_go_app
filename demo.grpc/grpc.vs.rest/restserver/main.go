package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"demo.grpc/grpc.vs.rest/pb"
)

var count int

func handle(w http.ResponseWriter, req *http.Request) {
	count++
	log.Println("rest access count: " + strconv.Itoa(count))

	decoder := json.NewDecoder(req.Body)
	var random pb.Random
	if err := decoder.Decode(&random); err != nil {
		panic(err)
	}

	random.RandomString = "[Updated] " + random.RandomString
	bytes, err := json.Marshal(&random)
	if err != nil {
		panic(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func main() {
	server := &http.Server{Addr: ":8080", Handler: http.HandlerFunc(handle)}
	log.Println("rest server start listen at: 8080")
	log.Fatal(server.ListenAndServeTLS("server.crt", "server.key"))
}
