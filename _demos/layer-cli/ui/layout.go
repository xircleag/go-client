package ui

import (
	"github.com/jroimartin/gocui"
)

type Dimensions struct {
	x  int
	x2 int
	y  int
	y2 int
}

var (
	HeaderViewOffset        = Dimensions{x: -1, y: -1}
	StatusViewOffset        = Dimensions{x: -1, y: 0}
	ConversationsViewOffset = Dimensions{x: -1, y: 9}
	ConversationViewOffset  = Dimensions{x: 41, y: 10}
	InputViewOffset         = Dimensions{x: 40}
	conversationCh          = make(chan bool)
)

func Layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if err := HeaderView(g, HeaderViewOffset.x, HeaderViewOffset.y, maxX, 1); err != nil {
		panic(err)
	}
	/*

		if err := StatusView(g, StatusViewOffset.x, StatusViewOffset.y, maxX, 10); err != nil {
			panic(err)
		}

		if err := ConversationsView(g, ConversationsViewOffset.x, ConversationsViewOffset.y, 40, maxY); err != nil {
			panic(err)
		}

		if err := InputView(g, 40, maxY-5, maxX-1, maxY-1); err != nil {
			panic(err)
		}
	*/

	/*
		if err := ContentView(g, ConversationViewOffset.x, ConversationViewOffset.y, maxX, maxY); err != nil {
			panic(err)
		}
	*/

	if err := MenuView(g, maxX, maxY); err != nil {
		panic(err)
	}

	//g.SetViewOnTop("status")
	g.SetViewOnTop("input")
	g.BgColor = gocui.Attribute(16 + 1)

	return nil
}
