package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"nsq-http-sub/http_api"
)

func OpHandler(w http.ResponseWriter, req *http.Request) {
	switch req.URL.Path {
	case "/channel/create":
		fallthrough
	case "/channel/pause":
		fallthrough
	case "/channel/unpause":
		fallthrough
	case "/channel/empty":
		fallthrough
	case "/channel/delete":
		err := broadcastChannelOp(req.URL.Path, req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write([]byte("OK"))
	default:
		w.WriteHeader(404)
	}

}

func broadcastChannelOp(op string, req *http.Request) (err error) {
	reqParams, err := http_api.NewReqParams(req)
	if err != nil {
		return
	}

	topicName, channelName, err := http_api.GetTopicChannelArgs(reqParams)
	if err != nil {
		return
	}

	for _, addr := range getAllProducersHttpAddr(topicName) {
		req_url := fmt.Sprintf("%s%s?topic=%s&channel=%s", addr, op, topicName, channelName)

		resp, err := http.PostForm(req_url, url.Values{})
		if err != nil {
			log.Printf("request nsqd fail %s %s", req_url, err.Error())
			return err
		}
		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		log.Printf("request nsqd %s [%s] %v", req_url, string(respBody), err)
	}

	return nil
}

func getAllProducersHttpAddr(topicName string) []string {
	list := []string{}

	for _, addr := range ShuffleStringArray(lookupdHTTPAddrs) {

		var url string
		if topicName == "" {
			url = "http://" + addr + "/nodes"
		} else {
			url = "http://" + addr + "/lookup?topic=" + topicName
		}

		resp, err := http.Get(url)

		if err != nil {
			log.Printf("ERROR with lookup topic=%s %s", topicName, err.Error())
			continue
		}
		defer resp.Body.Close()

		respBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("ERROR with lookup topic=%s %s", topicName, err.Error())
			continue
		}

		respStruct := &NSQLookupdResponse{}
		json.Unmarshal(respBody, respStruct)

		log.Printf("%#v", respStruct)

		if respStruct.StatusCode != 200 {
			continue
		}
		for _, producer := range respStruct.Data.Producers {
			log.Printf("lookup producer topic=%s %v", topicName, producer)
			list = append(list[:], fmt.Sprintf("http://%s:%d", producer.BroadcastAddress, producer.HttpPort))
		}

		break

	}

	return list
}

type NSQLookupdResponse struct {
	StatusCode int            `json:"status_code"`
	StatusTxt  string         `json:"status_txt"`
	Data       NSQLookupdData `json:data`
}

type NSQLookupdData struct {
	Channels  []string      `json:"channels"`
	Producers []NSQProducer `json:"producers"`
}

type NSQProducer struct {
	RemoteAddress    string `json:"remote_address"`
	HostName         string `json:"hostname"`
	BroadcastAddress string `json:"broadcast_address"`
	TcpPort          int    `json:"tcp_port"`
	HttpPort         int    `json:"http_port"`
	Version          string `json:"version"`
}
