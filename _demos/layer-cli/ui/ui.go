package ui

import (
	"fmt"
	"log"
	//"os"

	client "github.com/layerhq/go-client/_demos/layer-cli/client"

	"github.com/jroimartin/gocui"
	"github.com/urfave/cli"
)

var Client *client.Client

func Run(c *cli.Context) {
	g, err := gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	Client, err = client.NewClient(c)
	if err != nil {
		panic(err)
	}
	Client.GUI = g

	g.SetManagerFunc(Layout)
	g.Cursor = true
	g.Mouse = true

	if err = setupBindings(g); err != nil {
		log.Panicln(err)
	}

	// Start the client
	go Client.Start()

	// Run the main UI loop
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func selectConversation(g *gocui.Gui, v *gocui.View) error {
	if _, err := g.SetCurrentView(v.Name()); err != nil {
		return err
	}

	cx, cy := v.Cursor()
	if v, err := g.SetView("message", 50, 30, 70, 32); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, fmt.Sprintf("X: %d, Y: %d", cx, cy))
	}

	return nil
}

func newConversation(g *gocui.Gui, v *gocui.View) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("newConversation", maxX/2-20, maxY/2, maxX/2+20, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = "New Conversation"

		v.FgColor = gocui.Attribute(15 + 1)
		v.BgColor = gocui.Attribute(0)

		v.Autoscroll = false
		v.Editable = true
		v.Wrap = false
		v.Frame = true

		v.Editor = gocui.EditorFunc(inputEditor)

		_, err := g.SetCurrentView("newConversation")
		if err != nil {
			return err
		}
	}

	return nil
}

func setupBindings(g *gocui.Gui) error {
	// Key bindings
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone,
		func(g *gocui.Gui, v *gocui.View) error {
			return gocui.ErrQuit
		}); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlN, gocui.ModNone, newConversation); err != nil {
		return err
	}

	// Mouse bindings

	return nil
}
