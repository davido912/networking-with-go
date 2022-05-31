package main

import (
	"fmt"
	"github.com/jroimartin/gocui"
	"log"
	"time"
)

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.InputEsc = true

	g.SetManagerFunc(func(g *gocui.Gui) error {
		termwidth, termheight := g.Size()
		_, err := g.SetView("output", 0, 0, termwidth-1, termheight-4)
		if err != nil {
			return err
		}
		_, err = g.SetView("input", 0, termheight-3, termwidth-1, termheight-1)
		if err != nil {
			return err
		}
		return nil
	})

	// Terminal width and height.
	termwidth, termheight := g.Size()

	// Output.
	ov, err := g.SetView("output", 0, 0, termwidth-1, termheight-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create output view:", err)
		return
	}
	ov.Title = " Messages  -  <" + "main channel" + "> "
	ov.FgColor = gocui.ColorRed
	ov.Autoscroll = true
	ov.Wrap = true

	iv, err := g.SetView("input", 0, termheight-3, termwidth-1, termheight-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create input view:", err)
		return
	}
	iv.Title = " New Message  -  <" + "david" + "> "
	iv.FgColor = gocui.ColorCyan

	iv.Editable = true
	err = iv.SetCursor(0, 0)
	if err != nil {
		log.Println("Failed to set cursor:", err)
		return
	}
	go func() {
		//bs := make([]byte, 1024)
		for {
			time.Sleep(time.Second * 2)
			_, err := fmt.Fprintf(ov, iv.Buffer())
			if err != nil {
				panic(err)
			}
		}
	}()
	_, err = g.SetCurrentView("input")
	if err != nil {
		log.Println("Cannot set focus to input view:", err)
	}

	if err := g.SetKeybinding("", gocui.KeyEsc, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	//time.Sleep(time.Second * 5)
	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("hello", maxX/2-7, maxY/2, maxX/2+7, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		fmt.Fprintln(v, "Hello world!")
	}
	return nil
}

//func ent(g *gocui.Gui) error {
//	termwidth, termheight := g.Size()
//	fmt.Println(termwidth, termheight)
//
//	ov, err := g.SetView("output", 0, 0, termwidth-1, termheight-4)
//	if err != nil && err != gocui.ErrUnknownView {
//		log.Println("Failed to create output view:", err)
//		return err
//	}
//	ov.Title = " Messages  -  <" + "flower" + "> "
//	ov.FgColor = gocui.ColorRed
//	ov.Autoscroll = true
//	ov.Wrap = true
//
//	// Send a welcome message.
//	_, err = fmt.Fprintln(ov, "<Go-Chat>: Welcome to Go-Chat powered by PubNub!")
//	if err != nil {
//		log.Println("Failed to print into output view:", err)
//		return err
//	}
//	_, err = fmt.Fprintln(ov, "<Go-Chat>: Press Ctrl-C to quit.")
//	if err != nil {
//		log.Println("Failed to print into output view:", err)
//		return err
//	}
//	return err
//}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
