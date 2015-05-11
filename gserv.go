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

var topClientId = 0

type Server struct {
	mux *http.ServeMux
}

type Message struct {
	content string
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
	topClientId++
	return &Client{
		id:     topClientId,
		ws:     ws,
		server: server,
		msgCh:  make(chan *Message, msgChanSize),
		doneCh: make(chan bool),
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

func (c *Client) Listen() {
	enc := json.NewEncoder(c.ws)
	dec := json.NewDecoder(c.ws)

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
		case "cmd":
			err := sendClient(enc, "notice", "I don't know: "+m["data"])
			if err != nil {
				return
			}
		default:
			err := sendClient(enc, "notice", "INVALID TYPE: "+m["type"])
			if err != nil {
				return
			}
		}
	}
}

func NewServer(homeTemplate *template.Template) *Server {
	serv := &Server{
		mux: http.NewServeMux(),
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
	}))
	return serv
}

func (s *Server) Add(c *Client) {
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
