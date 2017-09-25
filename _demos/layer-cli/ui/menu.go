package ui

import (
	"bytes"
	"fmt"
	"time"

	"github.com/layerhq/go-client/_demos/layer-cli/helpers"

	"github.com/jroimartin/gocui"
)

// c - create new conversation
// l - list conversations
// p - change presence status
// s - settings
// q - exit

func MenuView(g *gocui.Gui, maxX, maxY int) error {

	if v, err := g.SetView("menu", -1, maxY-4, maxX, maxY+1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}

		_, err := g.SetCurrentView("menu")

		if err != nil {
			return err
		}

		v.FgColor = gocui.Attribute(15 + 1)
		v.BgColor = gocui.Attribute(0)

		v.Autoscroll = false
		v.Editable = false
		v.Wrap = false
		v.Frame = false

		go func() {
			for range time.Tick(time.Millisecond * 100) {
				UpdateMenuView(g, v)
			}
		}()
	}

	return nil
}

func UpdateMenuView(gui *gocui.Gui, v *gocui.View) {
	gui.Execute(func(g *gocui.Gui) error {
		v.Clear()
		v.SetCursor(0, 0)
		v.SetOrigin(0, 0)

		var conversationList = []string{}
		conversations := Client.Conversations()

		for _, conversation := range conversations {
			conversationList = append(conversationList, "")
			_ = conversation
		}

		activeConversation := Client.ActiveConversation
		activeConversationTitle := ""
		if activeConversation != nil {
			activeConversationTitle = activeConversation.Title()
		}
		timestamp := time.Now().Format("3:04:05 PM")
		buf := bytes.NewBufferString(fmt.Sprintf("⣿ %s ⡇ %d conversations ⡇ [%s]\n\n>",
			helpers.ColorStringf(10, timestamp),
			len(conversations),
			helpers.ColorStringf(12, activeConversationTitle)))

		//fmt.Fprintf(buf, "%s")

		v.Write(buf.Bytes())

		maxX, maxY := g.Size()

		if err := InputView(g, 1, maxY-2, maxX, maxY); err != nil {
			panic(err)
		}

		return nil
	})
}
