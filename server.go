package main

import (
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"

	"./message"
)

type ConnectionHandler struct {
	connections []net.Conn
	messages    []message.Message
	log         []string
}

func (this *ConnectionHandler) newConnection(c net.Conn) {
	this.connections = append(this.connections, c)
}

func (this *ConnectionHandler) removeConnection(c net.Conn) {
	for i, e := range this.connections {
		if e == c {
			this.connections = append(this.connections[:i], this.connections[i+1:]...)
			return
		}
	}
}

func (this *ConnectionHandler) newLog(log string) {
	this.log = append(this.log, log)
	updateUI()
}

func (this *ConnectionHandler) handleMessage(m message.Message) {
	this.messages = append(this.messages, m)
	if m.Type == "file" {
		err := ioutil.WriteFile(m.Body, m.File, 0644)
		if err != nil {
			this.newLog(err.Error())
		}
	}
	for _, c := range this.connections {
		err := gob.NewEncoder(c).Encode(m)
		if err != nil {
			this.newLog(err.Error())
		}
	}
	updateUI()
}

var chatRoom ConnectionHandler

func server() {
	s, err := net.Listen("tcp", "192.168.100.9:5400")
	if err != nil {
		return
	}
	for {
		c, err := s.Accept()
		if err == nil {
			go handleClient(c)
		} else {
			chatRoom.newLog(err.Error())
		}
	}
}

func handleClient(c net.Conn) {
	chatRoom.newConnection(c)

	var user string
	gob.NewDecoder(c).Decode(&user)
	chatRoom.newLog(user + " connected")

	var m message.Message
	for {
		err := gob.NewDecoder(c).Decode(&m)
		if err != nil {
			chatRoom.newLog(user + " disconnected")
			chatRoom.removeConnection(c)
			return
		}
		go chatRoom.handleMessage(m)
	}
}

func updateUI() {
	printHeader()
	for _, m := range chatRoom.log {
		fmt.Println("[" + m + "]")
	}
	fmt.Print("\n")
	for _, m := range chatRoom.messages {
		if m.Type == "text" {
			fmt.Println(m.User + ": " + m.Body)
		} else {
			fmt.Println(m.User + ": file[" + m.Body + "]")
		}
	}
	fmt.Print("\n > ")
}

func backup(path string) {
	file, err := os.Create(path)
	if err != nil {
		chatRoom.newLog(err.Error())
		return
	}

	for _, m := range chatRoom.messages {
		if m.Type == "text" {
			file.WriteString(m.User + ": " + m.Body + "\n")
		} else {
			file.WriteString(m.User + ": file[" + m.Body + "]\n")
		}
	}
	file.Close()
}

func printHeader() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println("Available commands:\n  'backup/<filename>' Back up the conversation\n  'exit' Ends server\n- - - - - - - - - -\n")
}

func main() {
	go server()

	var input string
	for input != ":exit" {
		updateUI()
		fmt.Scanln(&input)
		if input == "exit" {
			fmt.Println("Server off")
			return
		} else if len(input) > 7 && input[:6] == "backup" {
			backup(input[7:])
		} else {
			fmt.Println("Command unavailable(", input, ")")
		}
	}
}
