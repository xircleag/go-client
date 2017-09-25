package client

import (
	"bytes"
	"fmt"
	"strings"
	"sync"

	"github.com/layerhq/go-client/_demos/layer-cli/helpers"
	"github.com/layerhq/go-client/client"
	"github.com/layerhq/go-client/iterator"

	"github.com/jroimartin/gocui"
	"golang.org/x/net/context"
)

type RenderHandlerFunc func(*Conversation, *gocui.View) error

type Conversation struct {
	Unread        int
	Client        *Client
	Conversation  *client.Conversation
	RenderHandler RenderHandlerFunc

	MaxX int
	MaxY int

	mu       sync.Mutex
	rendered bool
}

// Create a title string from a list of conversation participants
func (convo *Conversation) Title() string {
	buf := bytes.NewBufferString("")
	for i, participant := range convo.Conversation.Participants {
		displayName := "Unknown User"
		if participant.DisplayName != "" {
			displayName = participant.DisplayName
		} else if participant.UserID != "" {
			displayName = participant.UserID
		}
		fmt.Fprintf(buf, "%s", displayName)
		if i < len(convo.Conversation.Participants)-1 {
			fmt.Fprintf(buf, ", ")
		}
	}
	return string(buf.Bytes())
}

// Render a conversation
func (convo *Conversation) Render(update bool) error {
	maxX, maxY := convo.Client.GUI.Size()
	v, err := convo.Client.GUI.SetView(convo.Conversation.ID, -1, 0, maxX, maxY-3)

	if err != gocui.ErrUnknownView {
		return err
	}

	// General layout
	v.Wrap = true
	v.Autoscroll = true
	v.Wrap = true
	// view.Highlight = true
	v.Frame = false

	v.FgColor = gocui.ColorWhite
	v.BgColor = gocui.Attribute(0)

	// Initial render options
	if !convo.rendered {
		fmt.Fprintf(v, helpers.ColorStringf(248, "\n\nConversation ID %s\n", convo.Conversation.ID))

		// Load the original messages
		var lines []string
		ctx := context.Background()
		messages, err := convo.Conversation.Messages(ctx)
		if err != nil {
			return err
		}
		for {
			message, err := messages.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return err
			}
			for _, part := range message.Parts {
				if part.MimeType == "text/plain" {
					displayName := message.Sender.UserID
					if message.Sender.DisplayName != "" {
						displayName = message.Sender.DisplayName
					}
					lineColor := 2
					if strings.HasSuffix(message.Sender.UserID, convo.Client.Username) {
						lineColor = 5
					}

					lines = append(lines, fmt.Sprintf("%s %s %s\n",
						helpers.ColorStringf(242, "%s", message.SentAt.Format("3:04:05 PM")),
						helpers.ColorStringf(lineColor, fmt.Sprintf("<%s>", displayName)),
						strings.TrimSpace(part.Body)))
				}
			}
		}

		// Reverse it
		last := len(lines) - 1
		for i := 0; i < len(lines)/2; i++ {
			lines[i], lines[last-i] = lines[last-i], lines[i]
		}
		for _, line := range lines {
			fmt.Fprintf(v, line)
		}

		convo.rendered = true
	}

	if !update {
		if err := convo.RenderHandler(convo, v); err != nil {
			return err
		}
	}

	return nil
}

func (convo *Conversation) View() (*gocui.View, error) {
	return convo.Client.GUI.View(convo.Conversation.ID)
}
