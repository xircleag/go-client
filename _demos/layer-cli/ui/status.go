package ui

import (
	"bytes"
	"fmt"
	"time"

	//"github.com/gizak/termui"
	"github.com/jroimartin/gocui"
)

var ()

func StatusView(g *gocui.Gui, x, y, maxX, maxY int) error {
	if v, err := g.SetView("status", x, y, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		v.FgColor = gocui.Attribute(15 + 1)
		v.BgColor = gocui.Attribute(0 + 1)

		v.Autoscroll = false
		v.Editable = false
		v.Wrap = false
		v.Frame = false
		v.Overwrite = true

		go func() {
			for range time.Tick(time.Millisecond * 100) {
				UpdateStatusView(g, v)
			}
		}()

	}

	return nil
}

func UpdateStatusView(g *gocui.Gui, v *gocui.View) {
	g.Execute(func(g *gocui.Gui) error {
		v.Clear()
		v.SetOrigin(0, 0)
		v.SetCursor(1, 1)

		buf := bytes.NewBufferString("\n\n")

		if Client.Connected {
			fmt.Fprintf(buf, " Status: \033[32;1mConnected\033[0m\n")
		} else {
			fmt.Fprintf(buf, " Status: \033[31;1mDisconnected\033[0m\n")
		}

		if len(Client.CounterLatency) > 1 {
			Client.StatsMu.RLock()
			lastCounter := Client.CounterLatency[len(Client.CounterLatency)-2]
			Client.StatsMu.RUnlock()
			lastDuration := lastCounter.End.Sub(lastCounter.Start)
			fmt.Fprintf(buf, " Last counter latency: %v", lastDuration)
		}

		v.Write(buf.Bytes())
		return nil
	})
}
