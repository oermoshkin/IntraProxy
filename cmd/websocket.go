package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"strings"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSConn struct {
	ClientConn *websocket.Conn
	ServerConn *websocket.Conn
}

func (w *WSConn) ClientRead() {
	for {
		mt, message, err := w.ClientConn.ReadMessage()
		if err != nil {
			w.Destroy()
			log.Println("Error read wss from client:", err)
			return
		}

		//log.Println("wss client read:", string(message))

		err = w.ServerWrite(mt, &message)
		if err != nil {
			w.Destroy()
			log.Println("Error write wss to server:", err)
			return
		}
	}
}

func (w *WSConn) ClientWrite(mt int, message *[]byte) error {
	//log.Println("wss client write:", string(*message))

	err := w.ClientConn.WriteMessage(mt, *message)
	if err != nil {
		return err
	}
	return nil
}

func (w *WSConn) ServerRead() {
	for {
		mt, message, err := w.ServerConn.ReadMessage()
		if err != nil {
			w.Destroy()
			log.Println("Error read wss from server:", err)
			return
		}
		//log.Println("wss server read:", string(message))

		err = w.ClientWrite(mt, &message)
		if err != nil {
			w.Destroy()
			log.Println("Error write wss to client:", err)
			return
		}
	}
}

func (w *WSConn) ServerWrite(mt int, message *[]byte) error {
	//log.Println("wss server write:", string(*message))
	err := w.ServerConn.WriteMessage(mt, *message)
	if err != nil {
		return err
	}
	return nil
}

func (w *WSConn) Destroy() {
	w.ServerConn.Close()
	w.ClientConn.Close()
}

func NewWS(w http.ResponseWriter, r *http.Request) {
	ClientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade failed: ", err)
		return
	}

	url := strings.Join([]string{"wss://", Config.Origin.ApiGW, r.URL.Path, "?", r.URL.RawQuery}, "")

	ServerConn, err := ClientWS(url, r.Header)
	if err != nil {
		log.Println("Error create wss with remote server:", err)
		return
	}

	NewClient := WSConn{
		ClientConn: ClientConn,
		ServerConn: ServerConn,
	}

	go NewClient.ClientRead()
	NewClient.ServerRead()
}

func ClientWS(url string, header http.Header) (*websocket.Conn, error) {
	//Удаляем заголовки, ибо они будут дублироваться - duplicate header not allowed
	header.Del("Sec-Websocket-Version")
	header.Del("Upgrade")
	header.Del("Connection")
	header.Del("Sec-Websocket-Key")
	header.Del("Sec-Websocket-Extensions")

	c, _, err := websocket.DefaultDialer.Dial(url, header)
	if err != nil {
		return nil, err
	}

	return c, nil
}
