package ui

import (
	"bytes"
	"fmt"
	//"time"

	client "github.com/layerhq/go-client/_demos/layer-cli/client"

	"github.com/jroimartin/gocui"
)

func ConversationsView(g *gocui.Gui, x, y, maxX, maxY int) error {
	if v, err := g.SetView("conversations", x, y, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.FgColor = gocui.Attribute(15 + 1)
		v.BgColor = gocui.Attribute(235 + 1)

		v.Highlight = true
		v.SelBgColor = gocui.Attribute(239 + 1)
		v.SelFgColor = gocui.ColorWhite

		v.Autoscroll = false
		v.Editable = false
		v.Wrap = false
		v.Frame = false
		v.Overwrite = true

		UpdateConversationsView(g, v)

		/*
			go func() {
				for range time.Tick(time.Millisecond * 100) {
					UpdateConversationsView(g, v)
				}
			}()
		*/
	}

	return nil
}

func UpdateConversationsView(g *gocui.Gui, v *gocui.View) {
	/*
		g.Execute(func(g *gocui.Gui) error {
			v.Clear()
			v.SetOrigin(0, 0)
			v.SetCursor(0, 0)

			buf := bytes.NewBufferString("")
			for _, convo := range Client.Conversations() {
				for i, participant := range convo.Participants {
					displayName := participant.DisplayName
					if displayName == "" {
						displayName = "Unknown User"
					}
					fmt.Fprintf(buf, "%s", displayName)
					if i < len(convo.Participants)-1 {
						fmt.Fprintf(buf, ", ")
					}
				}
				fmt.Fprintf(buf, "\n")
			}
			v.Write(buf.Bytes())

			return nil
		})
	*/
	Client.Execute("conversations", func(convo *client.Conversation, g *gocui.Gui, v *gocui.View, c *client.Client) error {
		v.Clear()
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)

		// List all conversations
		buf := bytes.NewBufferString("\n")
		conversations := c.Conversations()
		for _, convo := range conversations {
			fmt.Fprintf(buf, "%s\n", convo.Title())
		}
		v.Write(buf.Bytes())

		return nil
	})
}
