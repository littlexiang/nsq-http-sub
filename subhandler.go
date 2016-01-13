package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/littlexiang/nsq-http-sub/http_api"
	"github.com/nsqio/go-nsq"
)

func ConnectToNSQAndLookupd(r *nsq.Consumer, lookupd []string) error {
	for _, addrString := range lookupd {
		log.Printf("lookupd addr %s", addrString)
		err := r.ConnectToNSQLookupd(addrString)
		if err != nil {
			return err
		}
	}

	return nil
}

func SubHandler(w http.ResponseWriter, req *http.Request) {
	reqParams, err := http_api.NewReqParams(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	topicName, channelName, err := http_api.GetTopicChannelArgs(reqParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "httpserver doesn't support hijacking", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("nsq_httpsub go-nsq/%s", nsq.VERSION)
	cfg.MaxInFlight = *maxInFlight
	r, err := nsq.NewConsumer(topicName, channelName, cfg)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	r.SetLogger(log.New(os.Stderr, "", log.LstdFlags), nsq.LogLevelInfo)

	sr := &StreamReader{
		topic:       topicName,
		channel:     channelName,
		consumer:    r,
		req:         req,
		conn:        conn,
		bufrw:       bufrw, // TODO: latency writer
		connectTime: time.Now(),
	}
	streamServer.Set(sr)

	log.Printf("[%s] new connection", conn.RemoteAddr().String())
	bufrw.WriteString("HTTP/1.1 200 OK\r\nConnection: close\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n")
	bufrw.Flush()

	r.AddHandler(sr)

	// TODO: handle the error cases better (ie. at all :) )
	errors := ConnectToNSQAndLookupd(r, lookupdHTTPAddrs)
	log.Printf("connected to NSQ %v", errors)

	if *timeout > 0 {
		go func(sr *StreamReader) {
			timer := time.AfterFunc(time.Duration(*timeout)*time.Second, func() {
				log.Printf("timeout after %d seconds", *timeout)
				sr.consumer.Stop()
			})
			for {
				select {
				case <-sr.consumer.StopChan:
					timer.Stop()
					return
				}
			}
		}(sr)
	}

	go sr.HeartbeatLoop()

	// this read allows us to detect clients that disconnect
	go func(rw *bufio.ReadWriter) {
		b, err := rw.ReadByte()
		if err != nil {
			log.Printf("disconnect client %s", err.Error())
		} else {
			log.Printf("unexpected data on request socket (%c); closing", b)
		}
		sr.consumer.Stop()
	}(bufrw)
}

func (sr *StreamReader) HeartbeatLoop() {
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer func() {
		sr.conn.Close()
		heartbeatTicker.Stop()
		streamServer.Del(sr)
	}()
	for {
		select {
		case <-sr.consumer.StopChan:
			return
		case ts := <-heartbeatTicker.C:
			sr.Lock()
			sr.bufrw.WriteString(fmt.Sprintf("{\"_heartbeat_\":%d}\n", ts.Unix()))
			sr.bufrw.Flush()
			sr.Unlock()
		}
	}
}

func (sr *StreamReader) HandleMessage(message *nsq.Message) error {
	sr.Lock()
	defer sr.Unlock()

	if (*maxMessages > 0) && (sr.messageCount >= *maxMessages) {
		errMsg := fmt.Sprintf("maxMessages reached %d/%d", sr.messageCount, *maxMessages)
		log.Print(errMsg)
		message.RequeueWithoutBackoff(1)
		//		return errors.New(errMsg)
		return nil
	}

	sr.bufrw.Write(message.Body)
	sr.bufrw.WriteString("\n")
	sr.bufrw.Flush()
	sr.messageCount++

	if sr.messageCount == *maxMessages {
		sr.consumer.Stop()
	}

	go func() {
		atomic.AddUint64(&streamServer.messageCount, 1)
	}()

	return nil
}

type StreamServer struct {
	// 64bit atomic vars need to be first for proper alignment on 32bit platforms
	messageCount uint64

	sync.RWMutex // embed a r/w mutex
	clients      map[uint]*StreamReader
}

var StreamReaderCounter uint = 0

func (s *StreamServer) Set(sr *StreamReader) {
	s.Lock()
	defer s.Unlock()
	StreamReaderCounter++
	sr.id = StreamReaderCounter
	s.clients[sr.id] = sr
}

func (s *StreamServer) Del(sr *StreamReader) {
	s.Lock()
	defer s.Unlock()
	delete(s.clients, sr.id)
}

var streamServer *StreamServer

type StreamReader struct {
	sync.RWMutex // embed a r/w mutex
	topic        string
	channel      string
	consumer     *nsq.Consumer
	req          *http.Request
	conn         net.Conn
	bufrw        *bufio.ReadWriter
	connectTime  time.Time
	messageCount int
	id           uint
}

func StatsHandler(w http.ResponseWriter, req *http.Request) {
	totalMessages := atomic.LoadUint64(&streamServer.messageCount)
	io.WriteString(w, fmt.Sprintf("Total Messages: %d\n\n", totalMessages))

	now := time.Now()
	for _, sr := range streamServer.clients {
		duration := now.Sub(sr.connectTime).Seconds()
		secondsDuration := time.Duration(int64(duration)) * time.Second // turncate to the second

		io.WriteString(w, fmt.Sprintf("[%s] [%s : %s] connected: %s\n",
			sr.conn.RemoteAddr().String(),
			sr.topic,
			sr.channel,
			secondsDuration))
	}
}
