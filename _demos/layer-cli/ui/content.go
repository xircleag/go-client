package ui

import (
	"fmt"
	"time"

	"github.com/layerhq/go-client/client"

	"github.com/jroimartin/gocui"
)

func ContentView(g *gocui.Gui, x, y, maxX, maxY int) error {

	if v, err := g.SetView("content", x, y, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.FgColor = gocui.Attribute(15 + 1)
		v.BgColor = gocui.Attribute(233 + 1)

		v.Autoscroll = false
		v.Editable = false
		v.Wrap = false
		v.Frame = false
		v.Overwrite = true

		Client.Client.Websocket.HandleFunc(client.WebsocketMethodCounterRead, func(w *client.Websocket, p *client.WebsocketPacket) {
			g.Execute(func(g *gocui.Gui) error {
				fmt.Fprintf(v, fmt.Sprintf("Counter: %s\n", p.Body.(*client.WebsocketResponse).RequestID))
				return nil
			})
		})

		go func() {
			for range time.Tick(time.Millisecond * 50) {
				UpdateContentView(g)
			}
		}()

	}

	return nil
}

func UpdateContentView(g *gocui.Gui) {
}
