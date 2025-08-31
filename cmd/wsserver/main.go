package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	amqp "github.com/rabbitmq/amqp091-go"
)

// simple hub to broadcast messages to websocket clients via Server-Sent Events-like using WebSocket
// We'll implement a minimal websocket without external deps by using Gorilla is not present; so use the standard library's HTTP Hijacker is complex.
// Instead, we will implement a very simple Server-Sent Events endpoint at /ws/events which satisfies a realtime feed over HTTP.
// However, the issue explicitly asks for a websocket route. To keep dependencies minimal, we implement a very small websocket using the gorilla/websocket package which is already common. Add it to go.mod if not present.

// We avoid creating a separate package to keep changes minimal.

type hub struct {
	clients   map[*client]struct{}
	add       chan *client
	remove    chan *client
	broadcast chan []byte
}

type client struct {
	conn *ConnWrapper
	out  chan []byte
}

// ConnWrapper abstracts gorilla websocket to keep code simple if swapped.
// We'll rely on gorilla/websocket implementation.

var upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}

type ConnWrapper struct{ *websocket.Conn }

func newHub() *hub {
	return &hub{
		clients:   make(map[*client]struct{}),
		add:       make(chan *client),
		remove:    make(chan *client),
		broadcast: make(chan []byte, 100),
	}
}

func (h *hub) run() {
	for {
		select {
		case c := <-h.add:
			h.clients[c] = struct{}{}
		case c := <-h.remove:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.out)
				c.conn.Close()
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.out <- msg:
				default:
				}
			}
		}
	}
}

func (h *hub) wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	cw := &ConnWrapper{Conn: conn}
	c := &client{conn: cw, out: make(chan []byte, 32)}
	h.add <- c
	go func() {
		defer func() { h.remove <- c }()
		for {
			_, _, err := cw.ReadMessage() // keep reading to detect close
			if err != nil {
				return
			}
		}
	}()
	go func() {
		for msg := range c.out {
			_ = cw.WriteMessage(websocket.TextMessage, msg)
		}
	}()
}

func main() {
	// Config via env
	rabbitURL := getenv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/")
	queue := getenv("RABBITMQ_QUEUE", "logs.empresas")
	addr := getenv("WS_HTTP_ADDR", ":8090")

	h := newHub()
	go h.run()

	// Rabbit consumer
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		log.Fatalf("RabbitMQ dial: %v", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("RabbitMQ channel: %v", err)
	}
	_, err = ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Queue declare: %v", err)
	}
	msgs, err := ch.Consume(queue, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Consume: %v", err)
	}
	go func() {
		for m := range msgs {
			h.broadcast <- m.Body
		}
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/ws/events", h.wsHandler)
	server := &http.Server{Addr: addr, Handler: mux}

	go func() {
		log.Printf("WebSocket server listening on %s at /ws/events", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
	_ = ch.Close()
	_ = conn.Close()
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
