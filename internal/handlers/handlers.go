package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsJSONPayload)

var clients = make(map[WebSocketConnection]string)

//views is the jet view set
var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

//upgradeConnection is the websocket upgrader from gorilla/websockets
var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type WebSocketConnection struct {
	*websocket.Conn
}

//WsJSONResponse defines the response sent back from websocket
type WsJSONResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

//WsJSONPayload
type WsJSONPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json: " "`
}

//WSEndPoint upgrades connection to websocket
func WSEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Client Connected to endpoint")
	var response WsJSONResponse
	response.Message = `<em><small>Connected to Server </small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	}

	go ListenForWS(&conn)
}

//ListenForWS is listen all new connections and send it to channel
func ListenForWS(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error in Listening : ", fmt.Sprintf("%v", r))
		}
	}()

	var payload WsJSONPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			// Do nothing
		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

//ListenToWSChannel
func ListenToWSChannel() {
	var response WsJSONResponse
	for {
		e := <-wsChan
		// response.Action = "Got Here"
		// response.Message = fmt.Sprintf("Some Message, and action was %s", e.Action)
		//broadcastToAll(response)
		switch e.Action {
		case "username":
			// Get a list of all user and send it back to all broadcast
			clients[e.Conn] = e.Username
			users := getUserList()
			response.Action = "list_users"
			response.ConnectedUsers = users
			broadcastToAll(response)
		case "left":
			response.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			response.ConnectedUsers = users
			broadcastToAll(response)
		case "broadcast":
			response.Action = "broadcast"
			response.Message = fmt.Sprintf("<strong>%s</strong>: %s", e.Username, e.Message)
			broadcastToAll(response)
		}
	}

}

//broadcastToAll
func broadcastToAll(response WsJSONResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("Websocket error in Broadcasting :", err)
			_ = client.Close()
			delete(clients, client)
		}
	}
}

//getUserList returns all users connected to the server
func getUserList() []string {
	var userList []string
	for _, x := range clients {
		if x != "" {
			userList = append(userList, x)
		}
	}
	sort.Strings(userList)
	return userList
}

//Home render the home page
func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}

}

//renderPage renders the jet template
func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil

}
