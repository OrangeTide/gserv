// gserv
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"golang.org/x/net/websocket"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"
)

const (
	msgChanSize uint   = 2
	version     string = "gserv-0.1p0"
)

// TODO: need to provide this in a race-free way
var topClientId = 0

type Server struct {
	mux    *http.ServeMux
	client map[int]*Client
}

type Message struct {
	pkttype string
	data    string
}

type Client struct {
	id     int
	ws     *websocket.Conn
	server *Server
	msgCh  chan *Message
	doneCh chan bool // TODO: we could close the msgCh?
}

func NewClient(ws *websocket.Conn, server *Server) *Client {
	if ws == nil || server == nil {
		return nil
	}
	// TODO: need to provide this in a race-free way
	id := topClientId
	topClientId++
	return &Client{
		id:     id,
		ws:     ws,
		server: server,
		msgCh:  make(chan *Message, msgChanSize),
		doneCh: make(chan bool),
	}
}

func (c *Client) SendMsg(m *Message) {
	c.msgCh <- m
}

func (c *Client) msgEncoderLoop(enc *json.Encoder) {
	// TODO: need to terminate this when msgCh closes
	for {
		select {
		case msg := <-c.msgCh:
			if msg == nil {
				log.Println("msgChan was closed!")
				return
			}
			err := sendClient(enc, msg.pkttype, msg.data)
			if err != nil {
				c.Close()
			}
		}
	}
}

func sendClient(enc *json.Encoder, pkttype string, data string) error {
	err := enc.Encode(map[string]string{
		"type": pkttype,
		"data": data,
	})
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func (c *Client) Close() {
	// for this to work msgEncoderLoop() must terminate when it starts getting nil messages
	close(c.msgCh)
}

func (c *Client) Listen() {
	defer func() {
		c.Close()
	}()
	enc := json.NewEncoder(c.ws)
	dec := json.NewDecoder(c.ws)

	go c.msgEncoderLoop(enc)
	err := sendClient(enc, "notice", "Welcome to the System! ("+version+")")
	if err != nil {
		return
	}

	for {
		var m map[string]string
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			log.Println(err)
			break
		}
		log.Printf("%s: %s\n", m["type"], m["data"])

		// TODO: process command....
		switch m["type"] {
		case "say":
			c.server.Broadcast("Someone says '%s'", m["data"])
		default:
			err := sendClient(enc, "notice", "INVALID COMMAND: "+m["type"])
			if err != nil {
				return
			}
		}
	}
}

func NewServer(homeTemplate *template.Template) *Server {
	serv := &Server{
		mux:    http.NewServeMux(),
		client: make(map[int]*Client),
	}
	serv.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		homeTemplate.Execute(w, req.Host)
		log.Println("Request processed!")
	})
	serv.mux.HandleFunc("/version", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "Version: "+version)
		log.Println("version requested!")
	})
	serv.mux.Handle("/ws", websocket.Handler(func(ws *websocket.Conn) {
		defer func() {
			err := ws.Close()
			if err != nil {
				log.Println(err)
				// TODO: send err to a channel
			}
			log.Println("Websocket closed.")
		}()

		log.Printf("Request of %s\n", ws.LocalAddr().String())
		// TODO: Content-Type: text/event-stream
		// conf.Header.Set("Content-Type", "text/event-stream")
		client := NewClient(ws, serv)
		serv.Add(client)
		client.Listen()
		serv.Del(client)
	}))
	return serv
}

func (s *Server) Add(c *Client) {
	id := c.id
	s.client[id] = c
}

func (s *Server) Del(c *Client) {
	id := c.id
	delete(s.client, id)
}

func (s *Server) Broadcast(format string, a ...interface{}) {
	data := fmt.Sprintf(format, a...)
	var m *Message = &Message{
		pkttype: "msg",
		data:    data,
	}
	for _, c := range s.client {
		c.SendMsg(m)
	}
}

func main() {
	flag.Parse()
	fmt.Println(version)
	root := flag.String("root", ".", "path to root")
	homeTemplate := template.Must(template.ParseFiles(filepath.Join(*root, "home.html")))
	serv := NewServer(homeTemplate)
	err := http.ListenAndServe(":8080", serv.mux)
	if err != nil {
		log.Fatal("Could not start HTTP server" + err.Error())
	}
}
