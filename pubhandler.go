package main

import (
	"bytes"
	"log"
	"math/rand"
	"net/http"
	"nsq-http-sub/http_api"
	"time"
)

func PubHandler(w http.ResponseWriter, req *http.Request) {
	reqParams, err := http_api.NewReqParams(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	topicName, err := reqParams.Get("topic")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	data := reqParams.Body
	if len(data) == 0 {
		http.Error(w, "empty message payload", http.StatusBadRequest)
		return
	}

	for _, addr := range getAllProducersHttpAddr("") {
		resp, err := http.Post(addr+"/pub?topic="+topicName, "text/plain", bytes.NewReader(data))
		log.Println("pub to ", addr, topicName)
		if err != nil {
			log.Println("pub failed ", addr, topicName, err)
			//try next producer
			continue
		} else {
			resp.Body.Close()
			w.Write([]byte("OK"))
			return
		}
	}
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

func ShuffleStringArray(a []string) []string {
	rand.Seed(time.Now().UnixNano())
	a_copy := make([]string, len(a))

	copy(a_copy, a)
	for i := range a_copy {
		j := rand.Intn(i + 1)
		a_copy[i], a_copy[j] = a_copy[j], a_copy[i]
	}
	return a_copy
}
