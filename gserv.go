// gserv
package main

import (
	"io"
	"fmt"
	"flag"
	"net/http"
	"log"
	"path/filepath"
	"html/template"	
	"golang.org/x/net/websocket"
)

/*
type wsHandler struct {
}

// This function 
// implements the Handler interface for wsHandler struct.
func (ws wsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
		
}
*/

const (
	msgChanSize uint = 2
	version string = "gserv version 0.1p0"
)

var topClientId = 0

type Server struct {
	mux *http.ServeMux
}

type Message struct {
	content string
}
type Client struct {
	id int
	ws *websocket.Conn
	server *Server
	msgCh chan *Message
	doneCh chan bool // TODO: we could close the msgCh
}

func NewClient(ws *websocket.Conn, server *Server) *Client {
	if ws == nil || server == nil {
		return nil
	}
	topClientId++
	return &Client{
		id: topClientId,
		ws: ws,
		server: server, 
		msgCh: make(chan *Message, msgChanSize),
		doneCh: make(chan bool),
	}
}

func (c *Client) Listen() {
	io.Copy(c.ws, c.ws) // simple echo example
}

func NewServer(homeTemplate *template.Template) *Server {
	/*** SERVER ***/
	serv := &Server{
			mux: http.NewServeMux(),
	}
	serv.mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/" {
				http.NotFound(w, req)
				return
		}		
		homeTemplate.Execute(w, req.Host)
		// fmt.Fprintln(w, "Hello World!")
		log.Println("Request processed!")
	})	
	serv.mux.HandleFunc("/version", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")				
		fmt.Fprintln(w, "Version: " + version)
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
		} ()

		log.Println("Websocket started!")
		
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
	
	
	/*** Template ***/
	root := flag.String("root", ".", "path to root")
	homeTemplate := template.Must(template.ParseFiles(filepath.Join(*root, "home.html")))
	serv := NewServer(homeTemplate)
	err := http.ListenAndServe(":8080", serv.mux)
	if err != nil {
		log.Fatal("Could not start HTTP server", err)
	}	
}