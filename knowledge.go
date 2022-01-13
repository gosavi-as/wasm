package main

import (
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"
)

var (
	ws                    js.Value
	globalDoc             = js.Global().Get("document")
	webSocketRespCnt      = float32(0)
	webSocketRespTimeMS   = float32(0)
	domManipulationCnt    = float32(0)
	domManipulationTimeMS = float32(0)
)

/***
This is client side code.
This will go to client browser and execute over there.
***/

func registerCallbacks() {
	js.Global().Set("websSocketRequest", js.FuncOf(websSocketRequest))
}

/***
Used to establish websocket connection
*/
func webSocketConnection() {
	ws = js.Global().Get("WebSocket").New("ws://" + AddSrv + "/ws")

	// Call back for webSocket open
	ws.Call("addEventListener", "open", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		connResp := "WebSocket Connected Successfully with server!!!"
		println(connResp)
		ws.Call("send", connResp)
		globalDoc.Call("getElementById", "result").Set("value", connResp)
		return nil
	}))

	// Call back for webSocket close
	ws.Call("addEventListener", "close", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		connResp := "WebSocket Closed Successfully with server!!!"
		println(connResp)
		ws.Call("send", connResp)
		globalDoc.Call("getElementById", "result").Set("value", connResp)
		return nil
	}))

}

func webSocketEventListener() {
	// This part is to read response sent from server over websocket
	ws.Call("addEventListener", "message", js.FuncOf(func(this js.Value, args []js.Value) interface{} {
		t := time.Now().UnixNano()
		data := args[0].Get("data")
		var jsonMap map[string]interface{}
		err := json.Unmarshal([]byte(data.String()), &jsonMap)
		if err != nil {
			println("Error IS after unmarshalling:", err)
		}
		tUnixMilli := float32(time.Nanosecond) * float32(time.Now().UnixNano()-t) / float32(time.Millisecond)
		println("Websocket response time: ", tUnixMilli)
		webSocketRespCnt = webSocketRespCnt + 1
		webSocketRespTimeMS = webSocketRespTimeMS + tUnixMilli
		globalDoc.Call("getElementById", "wsResponseTime").Set("value", fmt.Sprintf("%f", webSocketRespTimeMS/webSocketRespCnt))

		// DOM Manipulation
		t = time.Now().UnixNano()
		fromContent := jsonMap["edges"].(map[string]interface{})

		// Manipulate table
		tableDom := "<tbody>"
		tableDom = "<tr><th scope=\"col\">Relation-1</th><th scope=\"col\">Relation-2</th><th scope=\"col\">Relation-3</th><th scope=\"col\">Relation-4</th><th scope=\"col\">Relation-5</th></tr>"
		for k, v := range fromContent {
			toContent := v.(map[string]interface{})
			tableDom += "<tr>"
			for k1, v1 := range toContent {
				relContent := v1.(map[string]interface{})
				for _, v2 := range relContent {
					tableDom += "<td>" + k + "->" + k1 + "->" + v2.(string) + "</td>"
				}
			}
			tableDom += "</tr>"
		}
		tableDom += "</tbody>"
		globalDoc.Call("getElementById", "resultTbl").Set("innerHTML", tableDom)
		tUnixMilli = float32(time.Nanosecond) * float32(time.Now().UnixNano()-t) / float32(time.Millisecond)
		println("Dom Manipulation time: ", tUnixMilli)
		domManipulationCnt = domManipulationCnt + 1
		domManipulationTimeMS = domManipulationTimeMS + tUnixMilli
		globalDoc.Call("getElementById", "domManTime").Set("value", fmt.Sprintf("%f", domManipulationTimeMS/domManipulationCnt))

		return nil
	}))
}

func webSocketInit() {
	webSocketConnection()
	webSocketEventListener()
}

/***
websSocketRequest This sends request to webSocketServer from client
*/
func websSocketRequest(this js.Value, i []js.Value) interface{} {
	val := globalDoc.Call("getElementById", i[0].String()).Get("value").String()
	println("Request: ", val)
	ws.Call("send", val)
	globalDoc.Call("getElementById", i[1].String()).Set("value", "Request: "+val)
	return nil
}

/***
WASM execution starts here on browser load
*/
func main() {
	c := make(chan struct{}, 0)
	println("CLIENT LOG: WASM Go Initialized at client", time.Now().String())

	//register functions
	registerCallbacks()
	// Websocket Initialization
	webSocketInit()

	<-c
}
