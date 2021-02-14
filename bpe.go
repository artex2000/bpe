package main

import (
	"github.com/Microsoft/go-winio"
	"github.com/gdamore/tcell"
	"log"
)

func main() {
	/*
	   err := InitWindow()
	   if err != nil {
	           return
	   }
	*/
	out, err := winio.DialPipe("\\\\.\\pipe\\console_log", nil)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer out.Close()

	log.SetOutput(out)

	win, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	err = win.Init()
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	defer win.Fini()

	win.Clear()
	x, y := win.Size()
	log.Printf("Screen size %v : %v\n", x, y)

	mouse := win.HasMouse()
	log.Printf("Screen mouse %v\n", mouse)

	colors := win.Colors()
	log.Printf("Screen colors %v\n", colors)

}
