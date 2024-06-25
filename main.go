// Copyright 2014 The gocui Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/jroimartin/gocui"
	"github.com/xiaoxiaoye/lazyssh/ssh"
)

var (
	hosts   = map[string]ssh.SSHConfig{}
	curHost string
)

func init() {
	sshConfigPath := os.ExpandEnv("$HOME/.ssh/config")
	configs, err := ssh.ParseSSHConfig(sshConfigPath)
	if err != nil {
		panic(err)
	}
	hosts = configs
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("side", -1, -1, int(0.2*float32(maxX)), maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Highlight = true
		v.SelBgColor = gocui.ColorGreen
		v.SelFgColor = gocui.ColorBlack
		for k := range hosts {
			fmt.Fprintln(v, k)
		}
		curHost = "host1"

		g.SetCurrentView("side")
	}
	if vm, err := g.SetView("main", int(0.2*float32(maxX)), -1, maxX, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		vm.Editable = true
		vm.Wrap = true
		vm.Overwrite = true
		if curHost != "" && vm != nil {
			vm.Clear()
			fmt.Fprintf(vm, "#Details: [%s]\n", hosts[curHost])
		}
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	return nil
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		writeHostDetail(g, v, cy+1)
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		writeHostDetail(g, v, cy-1)
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func writeHostDetail(g *gocui.Gui, v *gocui.View, y int) error {
	line, err := v.Line(y)
	if err != nil {
		line = ""
	}
	curHost = line
	vm, _ := g.View("main")
	if vm != nil {
		vm.Clear()
		fmt.Fprintf(vm, "##Details: [%s]\n", hosts[curHost])
	}
	return nil
}

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := keybindings(g); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
