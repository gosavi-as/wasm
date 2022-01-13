package main

import (
	"devrev/graph"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

const (
	AddSrv             = "localhost:8080"
	TemplateDir        = "."
	RandomLimit        = 500
	RelationLimit      = 5
	GraphRefreshTimeMS = 10000
)

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	GlobalGraph = graph.NewGraph()
)

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page")
}

/***
Request reader for requests coming from client
*/
func reader(conn *websocket.Conn) {
	for {
		_, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("ServerSideLog: " + string(p))
		requestAnalyser(conn, string(p))
	}
}

/***
Based on the request text generates response
*/
func requestAnalyser(conn *websocket.Conn, requestTxt string) {
	if requestTxt == "loadGraph" {
		graphBytes, err := json.Marshal(GlobalGraph)
		if err != nil {
			fmt.Errorf("%s", err)
		}
		fmt.Println("Writing GRAPH to WASM")
		socketWrite(conn, graphBytes)
	}
}

/***
Used to write response using webSocket
*/
func socketWrite(conn *websocket.Conn, msg []byte) {
	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Println(err)
		return
	}
}

/***
Upgrader to update from HTTP to WebSocket
*/
func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	log.Println("Client successfully connected....")
	go refreshGraph(ws)
	go reader(ws)
}

func setupRoutes() {
	fileSrv := http.FileServer(http.Dir(TemplateDir))
	http.Handle("/", fileSrv)
	http.HandleFunc("/ws", wsEndpoint)
}

/***
Function to randomly generate graph relationships
*/
func loadGraph() {
	GlobalGraph = graph.NewGraph()
	for i := 0; i < RandomLimit; i++ {
		from := "str:" + strconv.Itoa(i)
		toRandom := -1
		for j := 0; j < RelationLimit; j++ {
			for ok := true; ok; ok = i == toRandom {
				toRandom = rand.Intn(RandomLimit)
			}
			to := "str:" + strconv.Itoa(toRandom) // Because of this there could be fewer edges
			relation := "relation:" + strconv.Itoa(rand.Intn(RandomLimit))
			GlobalGraph.AddEdge(from, to, relation)
		}
	}
}

/***
Ticker to reload graph at configured interval
*/
func refreshGraph(conn *websocket.Conn) {
	ticker := time.NewTicker(GraphRefreshTimeMS * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				log.Println("Tick at", t)
				loadGraph()
				requestAnalyser(conn, "loadGraph")
			}
		}
	}()
}

func main() {
	log.Println("Starting server on Host:Port = ", AddSrv)
	setupRoutes()
	http.ListenAndServe(AddSrv, nil)
}
