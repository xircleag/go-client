package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

const (
	WebsocketMethodCounterRead        = "Method.Counter.read"
	WebsocketMethodConversationCreate = "Method.Conversation.create"
	WebsocketMethodMessageCreate      = "Method.Message.create"
	WebsocketMethodPresenceUpdate     = "Method.Presence.update"
	WebsocketMethodPresenceSync       = "Method.Presence.sync"

	WebsocketSignalTyping = "typing"

	WebsocketChangeConversationCreate          = "Change.ConversationCreate"
	WebsocektChangeConversationDelete          = "Change.ConversationDelete"
	WebsocketChangeConversationParticipants    = "Change.ConversationParticipants"
	WebsocketChangeConversationMetadata        = "Change.ConversationMetadata"
	WebsocketChangeConversationRecipientStatus = "Change.ConversationRecipientStatus"
	WebsocketChangeConversationLastMessage     = "Change.ConversationLastMessage"
	WebsocketChangeMessageCreate               = "Change.MessageCreate"
	WebsocektChangeMessageDelete               = "Change.MessageDelete"
)

type Websocket struct {
	client   *Client
	dialer   *websocket.Dialer
	conn     *websocket.Conn
	handlers *websocketEventHandlerSet
	wg       sync.WaitGroup
}

type WebsocketPacket struct {
	Type      string      `json:"type"`
	Body      interface{} `json:"body"`
	Counter   int         `json:"counter,omitempty"`
	Timestamp time.Time   `json:"-"`
}

type WebsocketChange struct {
	Operation string                `json:"operation"`
	Object    WebsocketChangeObject `json:"object"`
	Data      interface{}           `json:"data"`
}

type WebsocketChangeObject struct {
	Type string   `json:"type"`
	ID   string   `json:"id"`
	URL  *url.URL `json:"url"`
}

type WebsocketChangeData struct {
	Operation string `json:"operation"`
	Property  string `json:"property"`
	ID        string `json:"id"`
}

type WebsocketRequest struct {
	RequestID string      `json:"request_id"`
	Method    string      `json:"method"`
	ObjectID  string      `json:"object_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type WebsocketResponse struct {
	RequestID string      `json:"request_id"`
	Method    string      `json:"method"`
	ObjectID  string      `json:"object_id,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type WebsocketSignal struct {
}

// An interface to handle websocket event callbacks
type WebsocketEventHandler interface {
	Handle(w *Websocket, r *WebsocketResponse)
}

type WebsocketEventHandlerRemover interface {
	Remove()
}

type WebsocketHandlerFunc func(*Websocket, *WebsocketResponse)

func (hf WebsocketHandlerFunc) Handle(w *Websocket, r *WebsocketResponse) {
	hf(w, r)
}

type websocketEventHandlerSet struct {
	set map[string][]*websocketEventHandlerNode
	sync.RWMutex
}

type websocketEventHandlerNode struct {
	method  string
	set     *websocketEventHandlerSet
	handler WebsocketEventHandler
}

func (hn *websocketEventHandlerNode) Handle(w *Websocket, r *WebsocketResponse) {
	hn.handler.Handle(w, r)
}

func (hn *websocketEventHandlerNode) Remove() {
	hn.set.Lock()
	defer hn.set.Unlock()
	delete(hn.set.set, hn.method)
}

func newHandlerSet() *websocketEventHandlerSet {
	return &websocketEventHandlerSet{
		set: make(map[string][]*websocketEventHandlerNode),
	}
}

func (hs *websocketEventHandlerSet) add(method string, h WebsocketEventHandler) WebsocketEventHandlerRemover {
	method = strings.ToLower(method)
	hs.Lock()
	defer hs.Unlock()
	node := &websocketEventHandlerNode{
		method:  method,
		set:     hs,
		handler: h,
	}
	_, ok := hs.set[method]
	if !ok {
		hs.set[method] = []*websocketEventHandlerNode{node}
	} else {
		hs.set[method] = append(hs.set[method], node)
	}

	return node
}

// Dispatch events to registered handlers
func (hs *websocketEventHandlerSet) dispatch(w *Websocket, r *WebsocketResponse) {
	hs.Lock()
	defer hs.Unlock()

	method := strings.ToLower(r.Method)
	set, ok := hs.set[method]
	if !ok {
		return
	}

	wg := &sync.WaitGroup{}
	for _, h := range set {
		wg.Add(1)

		// Create a copy of the pointer
		hc := h
		go func() {
			hc.Handle(w, r)
			wg.Done()
		}()
	}
	wg.Wait()
}

func (w *Websocket) connect() error {
	// TODO: Add a mutex
	w.wg.Add(1)
	defer w.wg.Done()

	headers := http.Header{
		"Origin":                 {"http://local.host:80"},
		"Sec-WebSocket-Protocol": {"layer-1.0"},
	}

	w.dialer = &websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	token, err := w.client.transport.Session.Token()
	if err != nil {
		return err
	}

	u := fmt.Sprintf("%s?session_token=%s", w.client.websocketURL.String(), token)
	ws, _, err := w.dialer.Dial(u, headers)
	if err != nil {
		return err
	}
	w.conn = ws
	w.handlers.dispatch(w, &WebsocketResponse{Method: "connected"})

	return nil
}

// Send writes a websocket packet
func (w *Websocket) Send(ctx context.Context, p *WebsocketPacket) error {
	w.wg.Wait()

	if w.conn == nil {
		err := w.connect()
		if err != nil {
			return err
		}
	}

	wr, err := w.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}

	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	wr.Write(data)

	return wr.Close()
}

// Start listening for wbesocket events
func (w *Websocket) Listen(ctx context.Context) error {
	return w.Receive(ctx, func(ctx context.Context, p *WebsocketResponse) {
		// Dispatch
		if w.handlers != nil {
			w.handlers.dispatch(w, p)
		}
	})
}

func (w *Websocket) HandleFunc(method string, h WebsocketHandlerFunc) WebsocketEventHandlerRemover {
	if w.handlers == nil {
		w.handlers = &websocketEventHandlerSet{
			set: make(map[string][]*websocketEventHandlerNode),
		}
	}
	return w.handlers.add(method, h)
}

// Receive calls f with messages from the websocket
func (w *Websocket) Receive(ctx context.Context, f func(context.Context, *WebsocketResponse)) error {
	w.wg.Wait()

	if w.conn == nil {
		err := w.connect()
		if err != nil {
			return err
		}
	}

	for {
		// Create a response reader
		_, r, err := w.conn.NextReader()
		if err != nil {
			return err
		}

		res, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}

		// Decode the response
		var body json.RawMessage
		p := WebsocketPacket{Body: &body}
		if err := json.Unmarshal(res, &p); err != nil {
			return err
		}

		switch p.Type {
		case "response":
			var r *WebsocketResponse
			if err := json.Unmarshal(body, &r); err != nil {
				return err
			}
			f(ctx, r)
		}
	}

	return nil
}
