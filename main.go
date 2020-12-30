//Purpose: I wanted a simple and easy way to manage multiple terminals at
//the same time with the option of synchronization
//
//command line arguments:
//Number of Params: 1
//Content: location of login file
//login file row format:
//format: <pem file location> <domain location> <username>
//
//TUI Commands:
//-Use tab to cycle between connection views
//-Use 'loadCommand(<filename>)' to load a command from a file
//-Use Ctrl+s to toggle sync (commands running on all or only current terminal)
//-Use Ctrl+t to quick scroll to the top
//-Use Ctrl+b to quick scroll to the bottom
//-Use up and down arrow keys to scroll
//-Use Esc to quit and disconnect
package main

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/marcusolsson/tui-go"
	"github.com/marcusolsson/tui-go/wordwrap"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var cons []*ssh.Client
var views []tui.Widget
var helpview tui.Widget
var prefix []string

//parse file, format: <pem file location> <domain location> <username>
func loginSetup(fileloc string) {
	file, err := os.Open(fileloc)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner, count := bufio.NewScanner(file), 1
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), " ")
		pemBytes, err := ioutil.ReadFile(line[0]) //pem file location
		if err != nil {
			log.Fatal(err)
		}
		signer, err := ssh.ParsePrivateKey(pemBytes)
		if err != nil {
			log.Fatalf("parse key failed:%v", err)
		}
		config := &ssh.ClientConfig{
			User: line[2],
			Auth: []ssh.AuthMethod{ssh.PublicKeys(signer)},
			HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
				return nil
			},
		}
		conn, err := ssh.Dial("tcp", line[1]+":22", config)
		if err != nil {
			log.Fatalf("dial failed:%v", err)
		}
		cons = append(cons, conn)
		prefix = append(prefix, line[2]+"@"+line[1]+" ~ % ")
		v := tui.NewHBox(tui.NewLabel("Logged in for server " + strconv.Itoa(count)))
		views = append(views, v)
		count += 1
	}

	helpview = tui.NewHBox(tui.NewLabel(`Use tab to cycle between connection views
	Use 'loadCommand(<filename>)' to load a command from a file
	Use Ctrl+s to toggle sync (commands running on all or only current terminal)
	Use Ctrl+t to quick scroll to the top
	Use Ctrl+b to quick scroll to the bottom
	Use up and down arrow keys to scroll
	Use Esc to quit and disconnect
	Press tab to exit help view`))

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}

func loadCommand(inputtxt string) (string, bool, error) {
	res, resb, content := "", false, []byte{}
	matched, err := regexp.MatchString(`loadCommand(.*?)`, inputtxt)
	if err != nil {
		log.Fatal(err)
	}
	if matched {
		fileloc := strings.Replace(inputtxt, "loadCommand(", "", -1)
		fileloc = strings.Replace(fileloc, ")", "", -1)
		content, err = ioutil.ReadFile(fileloc)
		res = string(content)
		resb = true
	}
	return res, resb, err
}

func closeAll() {
	for i, _ := range cons {
		cons[i].Close()
	}
}

func cycle(viewnum int, max int) int {
	res := 0
	if !(viewnum+1 > max) {
		res = viewnum + 1
	}
	return res
}

func main() {
	argswithoutprog, synced := os.Args[1:], true

	loginSetup(argswithoutprog[0])
	defer closeAll()

	root, currentview := tui.NewVBox(views[0]), 0
	root.SetSizePolicy(tui.Maximum, tui.Maximum)

	rootScroll := tui.NewScrollArea(root)
	rootScroll.SetAutoscrollToBottom(true)
	rootScroll.SetSizePolicy(tui.Maximum, tui.Expanding)

	input := tui.NewEntry()
	input.SetFocused(true)
	input.SetSizePolicy(tui.Expanding, tui.Maximum)

	inputBox := tui.NewHBox(input)
	inputBox.SetBorder(true)
	inputBox.SetSizePolicy(tui.Expanding, tui.Maximum)

	chat := tui.NewVBox(rootScroll, inputBox)
	chat.SetSizePolicy(tui.Expanding, tui.Expanding)
	chat.SetBorder(true)

	ui, err := tui.New(chat)
	if err != nil {
		log.Fatal(err)
	}

	input.OnSubmit(func(e *tui.Entry) {
		root.Remove(0)
		inputtxt := e.Text()
		if val, isloadcommand, erri := loadCommand(inputtxt); isloadcommand {
			inputtxt = val
			if erri != nil {
				t := views[currentview]
				views[currentview] = tui.NewVBox(t, tui.NewHBox(tui.NewPadder(0, 0, tui.NewVBox(tui.NewLabel(wordwrap.WrapString(fmt.Sprintf("Error loading file name, please try again"), chat.Size().X-5)), tui.NewLabel("")))))
				root.Append(views[currentview])
				return
			}
		}
		input.SetText("Commands Running, please wait")
		if inputtxt == "help" {
			currentview--
			root.Remove(0)
			root.Append(helpview)
		} else if synced {
			for i, _ := range cons {
				var stdoutBuf bytes.Buffer
				sess, err := cons[i].NewSession()
				defer sess.Close()
				if err != nil {
					log.Fatalf("session failed:%v", err)
				}
				sess.Stdout = &stdoutBuf
				err = sess.Run(inputtxt + " 2>&1")
				if err != nil {
					stdoutBuf.WriteString(err.Error())
				}
				t := views[i]
				views[i] = tui.NewVBox(t, tui.NewHBox(tui.NewPadder(0, 0, tui.NewVBox(tui.NewLabel(wordwrap.WrapString(prefix[i]+inputtxt+"\n"+stdoutBuf.String(), chat.Size().X-5)), tui.NewLabel("")))))
				if i == currentview {
					root.Append(views[currentview])
				}
			}
		} else {
			var stdoutBuf bytes.Buffer
			sess, err := cons[currentview].NewSession()
			sess.Stdout = &stdoutBuf
			err = sess.Run(inputtxt + " 2>&1")
			if err != nil {
				stdoutBuf.WriteString(err.Error())
			}
			t := views[currentview]
			views[currentview] = tui.NewVBox(t, tui.NewHBox(tui.NewPadder(0, 0, tui.NewVBox(tui.NewLabel(wordwrap.WrapString(prefix[currentview]+inputtxt+"\n"+stdoutBuf.String(), chat.Size().X-5)), tui.NewLabel("")))))
			root.Append(views[currentview])
			sess.Close()
		}
		input.SetText("")
	})

	ui.SetKeybinding("Esc", func() { ui.Quit() })
	ui.SetKeybinding("TAB", func() {
		currentview = cycle(currentview, len(views)-1)
		root.Remove(0)
		root.Append(views[currentview])
	})
	ui.SetKeybinding("Ctrl+s", func() {
		synced = !synced
		root.Remove(0)
		t := views[currentview]
		views[currentview] = tui.NewVBox(t, tui.NewHBox(tui.NewPadder(0, 0, tui.NewVBox(tui.NewLabel(wordwrap.WrapString(fmt.Sprintf("switching synchronization to: %t", synced), chat.Size().X-5)), tui.NewLabel("")))))
		root.Append(views[currentview])
	})
	ui.SetKeybinding("Ctrl+b", func() {
		rootScroll.SetAutoscrollToBottom(true)
	}) 
	ui.SetKeybinding("Ctrl+t", func() {
		rootScroll.SetAutoscrollToBottom(false)
		rootScroll.ScrollToTop()
	})
	ui.SetKeybinding("Up", func() {
		rootScroll.SetAutoscrollToBottom(false)
		rootScroll.Scroll(0, -5)
	}) 
	ui.SetKeybinding("Down", func() {
		rootScroll.SetAutoscrollToBottom(false)
		rootScroll.Scroll(0, 5)
	})

	if err := ui.Run(); err != nil {
		log.Fatal(err)
	}
}
