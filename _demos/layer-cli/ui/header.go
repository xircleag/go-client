package ui

import (
	"bytes"
	"fmt"
	"time"

	"github.com/jroimartin/gocui"
)

func HeaderView(g *gocui.Gui, x, y, maxX, maxY int) error {

	if v, err := g.SetView("header", x, y, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.FgColor = gocui.Attribute(15 + 1)
		v.BgColor = gocui.Attribute(4 + 1)

		v.Autoscroll = false
		v.Editable = false
		v.Wrap = false
		v.Frame = false
		v.Overwrite = true

		fmt.Fprintf(v, "⣿ Loading...")

		go func() {
			for range time.Tick(time.Millisecond * 100) {
				UpdateHeaderView(g, v)
			}
		}()

	}

	return nil
}

func UpdateHeaderView(g *gocui.Gui, v *gocui.View) {
	g.Execute(func(g *gocui.Gui) error {
		v.Clear()
		v.SetOrigin(0, 0)
		v.SetCursor(0, 0)

		buf := bytes.NewBufferString("⣿ ")

		if Client.Connected {
			fmt.Fprintf(buf, fmt.Sprintf("\x1b[38;5;15mConnected\x1b[0m\x1b[38;5;117m to %s as %s\x1b[0m", Client.Client.BaseURL().String(), Client.Username))
		} else {
			fmt.Fprintf(buf, "\x1b[38;5;117mDisconnected\x1b[0m")
		}

		v.Write(buf.Bytes())
		return nil
	})
}
