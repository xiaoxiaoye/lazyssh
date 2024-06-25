package main

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/xiaoxiaoye/lazyssh/ssh"
	"golang.org/x/term"
)

func NewTerminal(config ssh.SSHConfig) {
	// c := exec.Command("bash")
	ip := config.Hostname
	port := config.Port
	if port == "" {
		port = "22"
	}
	user := config.User
	if user == "" {
		user = "root"
	}

	c := exec.Command("ssh", "-oPort="+port, user+"@"+ip)
	var err error
	ptmx, err := pty.Start(c)
	if err != nil {
		panic(err)
	}
	defer func() { _ = ptmx.Close() }() // Best effort.
	// ptmx.Write([]byte("export TERM=xterm-256color\n"))

	ws := pty.Winsize{Rows: 40, Cols: 120}
	pty.Setsize(ptmx, &ws)

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// Copy stdin to the pty and the pty to stdout.
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	_, _ = io.Copy(os.Stdout, ptmx)
}

func main() {
	app := tview.NewApplication()
	hosts, err := ssh.ParseSSHConfig()
	if err != nil {
		panic(err)
	}

	configView := tview.NewTextArea()
	configView.SetBorder(true).SetTitle("Config")

	hostView := tview.NewList()
	shortKey := rune('a')
	for host, config := range hosts {
		if shortKey == 'q' {
			shortKey++
		}
		copyH := config
		hostView.AddItem(host, "", shortKey, func() {
			app.Suspend(func() {
				NewTerminal(copyH)
			})
		})
		shortKey++
	}
	hostView.SetChangedFunc(func(i int, s string, s2 string, s3 rune) {
		config := hosts[s]
		sc, _ := json.MarshalIndent(config, "", "  ")
		configView.SetText(string(sc), true)
	})

	flex := tview.NewFlex().
		AddItem(hostView, 0, 1, true).AddItem(configView, 0, 2, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'q':
			app.Stop() // Quit the application
		}
		return event
	})
	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
