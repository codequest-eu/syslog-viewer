package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"

	"github.com/julienschmidt/httprouter"
	syslog "gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/mcuadros/go-syslog.v2/format"
)

type logEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Hostname  string    `json:"hostname"`
	Tag       string    `json:"tag"`
	Content   string    `json:"content"`
}

type logHandler struct {
	*syslog.ChannelHandler
	channel  syslog.LogPartsChannel
	clients  map[*websocket.Conn](chan []byte)
	quitChan chan struct{}
}

func newLogHandler() *logHandler {
	channel := make(syslog.LogPartsChannel)
	return &logHandler{
		syslog.NewChannelHandler(channel),
		channel,
		make(map[*websocket.Conn](chan []byte)),
		make(chan struct{}),
	}
}

func (lh *logHandler) keepProcessing() {
	for {
		select {
		case logParts := <-lh.channel:
			lh.processOne(logParts)
		case _ = <-lh.quitChan:
			break
		}
	}
}

func (lh *logHandler) processOne(logParts format.LogParts) {
	entry := logEntry{
		Timestamp: logParts["timestamp"].(time.Time),
		Hostname:  logParts["hostname"].(string),
		Tag:       logParts["tag"].(string),
		Content:   logParts["content"].(string),
	}
	data, err := json.Marshal(&entry)
	if err != nil {
		log.Printf("error encoding entry: %v", err)
		return
	}
	for _, dataChan := range lh.clients {
		dataChan <- data
	}
}

func (lh *logHandler) ws(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	websocket.Handler(lh.handleWebsocket).ServeHTTP(w, r)
}

func (lh *logHandler) handleWebsocket(conn *websocket.Conn) {
	dataChan := make(chan []byte)
	lh.clients[conn] = dataChan
	defer close(dataChan)
	defer delete(lh.clients, conn)
	for data := range dataChan {
		if _, err := conn.Write(data); err != nil {
			break
		}
	}
}

func (lh *logHandler) quit() {
	log.Print("Goodbye!")
	lh.quitChan <- struct{}{}
}

func indexHTML(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	http.ServeFile(w, r, "/frontend/index.html")
}

func main() {
	router := httprouter.New()
	router.GET("/", indexHTML)
	router.ServeFiles("/assets/*filepath", http.Dir("/frontend/assets"))
	handler := newLogHandler()
	defer handler.quit()
	router.GET("/ws", handler.ws)
	server := syslog.NewServer()
	server.SetFormat(syslog.RFC3164)
	server.SetHandler(handler)
	server.ListenUDP("0.0.0.0:514")
	server.Boot()
	go handler.keepProcessing()
	log.Print("Server accepting UDP on port 514 and HTTP on port 10514")
	log.Fatal(http.ListenAndServe(":10514", router))
}
