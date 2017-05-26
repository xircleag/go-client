package client

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"golang.org/x/net/context"
)

type Websocket struct {
	client *Client
	dialer *websocket.Dialer
	conn   *websocket.Conn
	wg     sync.WaitGroup
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
