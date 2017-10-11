package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
	"github.com/buger/jsonparser"
)

const (
	WebsocketMethodCounterRead        = "Counter.read"
	WebsocketMethodConversationCreate = "Conversation.create"
	WebsocketMethodMessageCreate      = "Message.create"
	WebsocketMethodPresenceUpdate     = "Presence.update"
	WebsocketMethodPresenceSync       = "Presence.sync"

	WebsocketSignalTyping = "typing"

	WebsocketChangeConversationCreate          = "Conversation.create"
	WebsocketChangeConversationDelete          = "Conversation.delete"
	WebsocketChangeConversationParticipants    = "Conversation.participants"
	WebsocketChangeConversationMetadata        = "Conversation.metadata"
	WebsocketChangeConversationMarkAllRead	   = "Conversation.mark_all_read"
	WebsocketChangeConversationRecipientStatus = "Conversation.recipient_status"
	WebsocketChangeConversationLastMessage     = "Conversation.last_message"
	WebsocketChangeMessageCreate               = "Message.create"
	WebsocketChangeMessageDelete               = "Message.delete"
)

type Websocket struct {
	client   *WebsocketClient
	dialer   *websocket.Dialer
	conn     *websocket.Conn
	handlers *websocketEventHandlerSet
	sync.RWMutex
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
	Type string `json:"type"`
	ID   string `json:"id"`
	URL  string `json:"url"`
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
	Handle(w *Websocket, r *WebsocketPacket)
}

type WebsocketEventHandlerRemover interface {
	Remove()
}

type WebsocketHandlerFunc func(*Websocket, *WebsocketPacket)

func (hf WebsocketHandlerFunc) Handle(w *Websocket, p *WebsocketPacket) {
	hf(w, p)
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

func (hn *websocketEventHandlerNode) Handle(w *Websocket, p *WebsocketPacket) {
	hn.handler.Handle(w, p)
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
func (hs *websocketEventHandlerSet) dispatch(w *Websocket, p *WebsocketPacket) {
	hs.Lock()
	defer hs.Unlock()

	handlerName := ""
	switch p.Body.(type) {
	case *WebsocketResponse:
		r := p.Body.(*WebsocketResponse)
		handlerName = strings.ToLower(r.Method)
	case *WebsocketChange:
		c := p.Body.(*WebsocketChange)
		handlerName = strings.ToLower(fmt.Sprintf("%s.%s", c.Object.Type, c.Operation))
	default:
		handlerName = "Unknown"
	}

	set, ok := hs.set[handlerName]
	if !ok {
		return
	}

	wg := &sync.WaitGroup{}
	for _, h := range set {
		wg.Add(1)

		// Create a copy of the pointer
		hc := h
		go func() {
			hc.Handle(w, p)
			wg.Done()
		}()
	}
	wg.Wait()
}

var wsHeaders = http.Header{
	"Origin":                 {"http://local.host:80"},
	"Sec-WebSocket-Protocol": {"layer-3.0"},
}

func (w *Websocket) connect() error {
	if w.conn != nil {
		return nil
	}

	w.Lock()
	defer w.Unlock()
	w.dialer = &websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	token, err := w.client.transport.Session.Token()
	if err != nil {
		return err
	}

	u := fmt.Sprintf("%s?session_token=%s", w.client.baseURL.String(), token)
	ws, _, err := w.dialer.Dial(u, wsHeaders)
	if err != nil {
		log.Println("error dialing " + err.Error())
		return err
	}
	w.conn = ws

	// Dispatch a connected event
	w.handlers.dispatch(w, &WebsocketPacket{
		Body: &WebsocketResponse{Method: "connected"},
	})

	return nil
}

// Send writes a websocket packet
func (w *Websocket) Send(ctx context.Context, p *WebsocketPacket) error {
	if err := w.connect(); err != nil {
		return err
	}

	w.RLock()
	wr, err := w.conn.NextWriter(websocket.TextMessage)
	if err != nil {
		return err
	}
	w.RUnlock()

	data, err := json.Marshal(p)
	if err != nil {
		return err
	}
	wr.Write(data)

	return wr.Close()
}

// Start listening for wbesocket events
func (w *Websocket) Listen(ctx context.Context) error {
	return w.Receive(ctx, func(ctx context.Context, p *WebsocketPacket) {
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

// Receive calls f with messages from the websocket, note this blocks until an error is encountered
func (w *Websocket) Receive(ctx context.Context, f func(context.Context, *WebsocketPacket)) error {
	if err := w.connect(); err != nil {
		return err
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
		p := &WebsocketPacket{Body: &body}
		if err := json.Unmarshal(res, &p); err != nil {
			return err
		}

		switch strings.ToLower(p.Type) {
		case "response":
			var data json.RawMessage
			r := &WebsocketResponse{Data: &data}
			if err := json.Unmarshal(body, &r); err != nil {
				return err
			}
			p.Body = r

			id, err := jsonparser.GetString(*r.Data.(*json.RawMessage), "id")
			objectType := strings.ToLower(id[9:])
			switch {
			case strings.HasPrefix(objectType, "conversations"):
				var conversation *Conversation
				if err = json.Unmarshal(*r.Data.(*json.RawMessage), &conversation); err == nil {
					r.Data = conversation
				}
			case strings.HasPrefix(objectType, "messages"):
				var message *Message
				if err = json.Unmarshal(*r.Data.(*json.RawMessage), &message); err == nil {
					r.Data = message
				}
			}
		case "change":
			var c *WebsocketChange
			if err := json.Unmarshal(body, &c); err != nil {
				return err
			}
			p.Body = c
			objectJSON, err := json.Marshal(c.Data)
			if err == nil {
				switch strings.ToLower(c.Object.Type) {
				case "conversation":
					var conversation *Conversation
					if err = json.Unmarshal(objectJSON, &conversation); err == nil {
						c.Data = conversation
					}
				case "message":
					var message *Message
					if err = json.Unmarshal(objectJSON, &message); err == nil {
						c.Data = message
					}
				}
			}
		}
		f(ctx, p)
	}

	return nil
}
