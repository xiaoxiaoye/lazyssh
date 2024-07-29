package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"syscall"

	"github.com/creack/pty"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	lazyssh "github.com/xiaoxiaoye/lazyssh/ssh"
	"golang.org/x/term"
)

func NewTerminal(conf lazyssh.SSHConfig) {
	// 创建执行命令
	ip := conf.Hostname
	port := conf.Port
	if port == "" {
		port = "22"
	}
	user := conf.User
	if user == "" {
		user = "root"
	}
	c := exec.Command("ssh", "-oPort="+port, user+"@"+ip)

	// 创建伪终端
	ptmx, err := pty.Start(c)
	if err != nil {
		fmt.Printf("Error starting pty: %v\n", err)
		return
	}
	defer func() { _ = ptmx.Close() }() // 确保程序退出时关闭伪终端

	// Set stdin in raw mode.
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }() // Best effort.

	// 捕获窗口大小变化信号
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func(tty *os.File) {
		for range ch {
			if err := pty.InheritSize(os.Stdin, tty); err != nil {
				fmt.Printf("Error resizing pty: %v\n", err)
			}
		}
	}(ptmx)
	ch <- syscall.SIGWINCH // 初始化窗口大小

	// 将标准输入、输出、错误连接到伪终端
	go func() { _, _ = io.Copy(ptmx, os.Stdin) }()
	go func() { _, _ = io.Copy(os.Stdout, ptmx) }()

	// 等待命令退出
	if err := c.Wait(); err != nil {
		fmt.Printf("Error waiting for command: %v\n", err)
	}
}

func main() {
	app := tview.NewApplication()
	setTheme()

	hostsConfig, err := lazyssh.ParseSSHConfig()
	if err != nil {
		panic(err)
	}

	hostsList := make([]lazyssh.SSHConfig, 0, len(hostsConfig))
	for _, config := range hostsConfig {
		hostsList = append(hostsList, config)
	}
	sort.Slice(hostsList, func(i, j int) bool {
		return hostsList[i].Host < hostsList[j].Host
	})

	configView := tview.NewTextArea()
	configView.SetBorder(true).SetTitle("Config")

	hostView := tview.NewList()
	shortKey := rune('a')
	for _, config := range hostsList {
		if shortKey == 'q' || shortKey == 'j' || shortKey == 'k' {
			shortKey++
		}
		copyH := config
		hostView.AddItem(config.Host, "", shortKey, func() {
			app.Suspend(func() {
				NewTerminal(copyH)
			})
		})
		shortKey++
	}
	hostView.SetChangedFunc(func(i int, s string, s2 string, s3 rune) {
		config := hostsConfig[s]
		sc, _ := json.MarshalIndent(config, "", "  ")
		configView.SetText(string(sc), true)
	})
	hostView.SetBorder(true).SetTitle("Hosts")
	hostView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Rune() {
		case 'j':
			hostView.SetCurrentItem((hostView.GetCurrentItem() + 1) % hostView.GetItemCount())
		case 'k':
			hostView.SetCurrentItem((hostView.GetCurrentItem() - 1 + hostView.GetItemCount()) % hostView.GetItemCount())
		}
		return event
	})

	flex := tview.NewFlex().
		AddItem(hostView, 0, 1, true).
		AddItem(configView, 0, 2, false)

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

func setTheme() {
	tview.Styles.PrimitiveBackgroundColor = tcell.ColorWhiteSmoke
	tview.Styles.TitleColor = tcell.ColorBlack
	tview.Styles.BorderColor = tcell.ColorBlack
	tview.Styles.PrimaryTextColor = tcell.ColorBlack
	tview.Styles.SecondaryTextColor = tcell.ColorBlack
}
