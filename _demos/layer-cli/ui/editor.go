package ui

import (
	"fmt"
	"strings"

	"github.com/layerhq/go-client/_demos/layer-cli/client"
	"github.com/layerhq/go-client/_demos/layer-cli/helpers"

	"github.com/jroimartin/gocui"
	"golang.org/x/net/context"
)

func inputEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	switch {
	case ch != 0 && mod == 0:
		v.EditWrite(ch)
	case key == gocui.KeySpace:
		v.EditWrite(' ')
	case key == gocui.KeyBackspace || key == gocui.KeyBackspace2:
		v.EditDelete(true)
	case key == gocui.KeyDelete:
		v.EditDelete(false)
	case key == gocui.KeyInsert:
		v.Overwrite = !v.Overwrite
	case key == gocui.KeyEnter:
		line := v.Buffer()

		v.Editable = false
		v.Clear()
		v.SetCursor(0, 0)
		v.SetOrigin(0, 0)

		if Client.GUI.CurrentView().Name() == "newConversation" {
			fmt.Fprintf(v, helpers.ColorStringf(242, "Creating..."))

			// Create a new conversation
			ctx := context.Background()
			newConversation, err := Client.Client.CreateConversation(ctx, []string{strings.TrimSpace(line)}, false, nil)
			if err != nil {
				return
			}

			// See if this conversation already exists
			conversations := Client.Conversations()
			for _, conversation := range conversations {
				if conversation.Conversation.ID == newConversation.ID {
					Client.ActiveConversation = conversation
					Client.ActiveConversation.Render(false)
					return
				}
			}

			// This is a new conversation
			Client.AddConversation(&client.Conversation{
				Client:       Client,
				Conversation: newConversation,
				RenderHandler: func(conversation *client.Conversation, view *gocui.View) error {
					return nil
				},
			})

			Client.Execute(newConversation.ID, func(convo *client.Conversation, g *gocui.Gui, cv *gocui.View, c *client.Client) error {
				return nil
			})
		} else {
			fmt.Fprintf(v, helpers.ColorStringf(242, "Sending..."))

			// Send message
			Client.Execute(Client.ActiveConversation.Conversation.ID, func(convo *client.Conversation, g *gocui.Gui, cv *gocui.View, c *client.Client) error {
				ctx := context.Background()
				c.ActiveConversation.Conversation.SendTextMessage(ctx, line, nil)
				v.Clear()
				v.SetCursor(0, 0)
				v.SetOrigin(0, 0)
				v.Editable = true
				return nil
			})
		}
	case key == gocui.KeyArrowDown:
		v.MoveCursor(0, 1, false)
	case key == gocui.KeyArrowUp:
		v.MoveCursor(0, -1, false)
	case key == gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case key == gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	}
}
