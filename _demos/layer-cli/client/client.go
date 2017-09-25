package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/layerhq/go-client/_demos/layer-cli/helpers"
	"github.com/layerhq/go-client/client"
	"github.com/layerhq/go-client/common"
	"github.com/layerhq/go-client/iterator"
	"github.com/layerhq/go-client/option"

	"github.com/jroimartin/gocui"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

var (
	maxTimings  = 10
	layoutViews = []string{"header", "conversations", "menu", "status"}
)

type Handler func(*Conversation, *gocui.Gui, *gocui.View, *Client) error

type Client struct {
	GUI                *gocui.Gui
	Client             *client.Client
	Connected          bool
	Username           string
	Identity           *client.BasicIdentity
	StatsMu            *sync.RWMutex
	CounterLatency     []timedRequest
	ActiveConversation *Conversation

	requestID     int
	conversations []*Conversation

	mu        *sync.Mutex
	convoMu   *sync.RWMutex
	contentMu *sync.RWMutex
}

type timedRequest struct {
	Start     time.Time
	End       time.Time
	RequestID string
}

type Credentials struct {
	ApplicationID string          `json:"application_id"`
	ProviderID    string          `json:"provider_id"`
	AccountID     string          `json:"account_id"`
	Key           *CredentialsKey `json:"key"`
}

type CredentialsKey struct {
	ID      string `json:"id"`
	Private string `json:"private"`
	Public  string `json:"public"`
}

func configureClient(username string, path string) (*client.Client, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c *Credentials
	json.Unmarshal(data, &c)

	ctx := context.Background()
	return client.NewClient(ctx, c.ApplicationID, option.WithCredentials(&common.ClientCredentials{
		User:       username,
		ProviderID: c.ProviderID,
		AccountID:  c.AccountID,
		Key: &common.Key{
			ID: c.Key.ID,
			KeyPair: &common.KeyPair{
				Private: c.Key.Private,
			},
		},
	}))
}

func NewClient(cc *cli.Context) (*Client, error) {
	layerClient, err := configureClient(cc.GlobalString("username"), cc.GlobalString("credentials-file"))
	if err != nil {
		return nil, err
	}
	//fmt.Println(fmt.Sprintf("%v", layerClient))

	return &Client{
		Client:   layerClient,
		Username: cc.GlobalString("username"),
		mu:       &sync.Mutex{},
		convoMu:  &sync.RWMutex{},
		StatsMu:  &sync.RWMutex{},
	}, nil
}

func (c *Client) AddConversation(conversation *Conversation) {
	c.convoMu.Lock()
	defer c.convoMu.Unlock()
	c.conversations = append(c.conversations, conversation)
}

func (c *Client) getConversations() {
	ctx := context.Background()
	convos, err := c.Client.Conversations(ctx, "")
	if err == nil {
		for {
			convo, err := convos.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				break
			}
			c.AddConversation(&Conversation{
				Client:       c,
				Conversation: convo,
				RenderHandler: func(conversation *Conversation, view *gocui.View) error {
					return nil
				},
			})
		}
	}
}

func (c *Client) Conversations() []*Conversation {
	c.convoMu.Lock()
	defer c.convoMu.Unlock()
	return c.conversations
}

func (c *Client) Execute(id string, handler Handler) {
	c.GUI.Execute(func(g *gocui.Gui) error {
		// Check if this is a core layout view
		for _, viewName := range layoutViews {
			if viewName == id {
				view, err := g.View(id)
				if err != nil {
					return err
				}
				return handler(nil, c.GUI, view, c)
			}
		}

		// See if this is an existing conversation view
		convo, _, exists := c.FindConversation(id)
		if convo != nil {
			view, err := convo.View()
			if err != nil {
				return err
			}
			return handler(convo, c.GUI, view, c)
		}
		_ = exists

		// Create a new conversation view
		fmt.Println(fmt.Sprintf("No view exists for the active conversation ID %s", id))
		os.Exit(1)
		return nil
	})
}

func (c *Client) FindConversation(id string) (*Conversation, int, error) {
	for i, convo := range c.conversations {
		if convo.Conversation.ID == id {
			return convo, i, nil
		}
	}
	return nil, -1, errors.New("Conversation not found")
}

func (c *Client) SetActiveConversation(offset int) {
	c.convoMu.RLock()
	defer c.convoMu.RUnlock()
	if offset <= len(c.conversations)-1 {
		c.ActiveConversation = c.conversations[offset]
		c.ActiveConversation.Render(false)
	}
}

func (c *Client) Start() {
	// Setup handlers
	c.Client.Websocket.HandleFunc("connected", func(w *client.Websocket, p *client.WebsocketPacket) {
		c.Connected = true
	})

	// Websocket counter
	c.Client.Websocket.HandleFunc(client.WebsocketMethodCounterRead, func(w *client.Websocket, p *client.WebsocketPacket) {
		end := time.Now()
		r := p.Body.(*client.WebsocketResponse)
		c.StatsMu.RLock()
		defer c.StatsMu.RUnlock()
		for i, val := range c.CounterLatency {
			if val.RequestID == r.RequestID {
				c.CounterLatency[i].End = end
				break
			}
		}
		if len(c.CounterLatency) > maxTimings {
			c.CounterLatency = c.CounterLatency[(len(c.CounterLatency) - maxTimings):]
		}
	})

	// Websocket events
	c.Client.Websocket.HandleFunc(client.WebsocketChangeMessageCreate, func(w *client.Websocket, p *client.WebsocketPacket) {
		change := p.Body.(*client.WebsocketChange)
		message := change.Data.(*client.Message)
		if message == nil {
			return
		}
		conversation := message.Conversation
		c.Execute(conversation.ID, func(convo *Conversation, g *gocui.Gui, v *gocui.View, c *Client) error {
			for _, part := range message.Parts {
				if part.MimeType == "text/plain" {
					displayName := "Unknown User"
					if message.Sender.DisplayName != "" {
						displayName = message.Sender.DisplayName
					}
					fmt.Fprintf(v, "%s %+v %+s %s",
						helpers.ColorStringf(242, "%s", message.SentAt.Format("3:04:05 PM")),
						message.Sender,
						helpers.ColorStringf(5, fmt.Sprintf("<%s>", displayName)),
						part.Body)
				}
			}
			return nil
		})
	})

	// Get the initial conversations
	c.getConversations()
	c.SetActiveConversation(0)

	// Start the client
	ctx := context.Background()
	go c.Client.Websocket.Listen(ctx)

	// Every 3 seconds, send a websocket counter request
	ctx, _ = context.WithTimeout(context.Background(), 3*time.Second)
	go func() {
		for range time.Tick(time.Second * 3) {
			requestID := fmt.Sprintf("%d", c.requestID)
			c.StatsMu.Lock()
			c.CounterLatency = append(c.CounterLatency, timedRequest{RequestID: requestID, Start: time.Now()})
			c.StatsMu.Unlock()
			c.Client.Websocket.Send(ctx, &client.WebsocketPacket{
				Type: "request",
				Body: &client.WebsocketRequest{
					Method:    client.WebsocketMethodCounterRead,
					RequestID: requestID,
				},
			})
			c.requestID += 1
		}
	}()
}
