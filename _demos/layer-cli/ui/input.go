package ui

import (
	"github.com/jroimartin/gocui"
)

func InputView(g *gocui.Gui, x, y, maxX, maxY int) error {
	if v, err := g.SetView("input", x, y, maxX, maxY); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		if g.CurrentView().Name() != "newConversation" {
			_, err := g.SetCurrentView("input")
			if err != nil {
				return err
			}
		}

		v.FgColor = gocui.Attribute(15 + 1)
		v.BgColor = gocui.Attribute(0)

		v.Autoscroll = false
		v.Editable = true
		v.Wrap = false
		v.Frame = false

		v.Editor = gocui.EditorFunc(inputEditor)
	}

	return nil
}
