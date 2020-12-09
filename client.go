package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"os/exec"

	"./message"
)

var messages = make([]message.Message, 0)

func printHeader() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
	fmt.Println("Available commands:\n  ':file/<filename>' Send file\n  ':download/<filename>' Download file\n  ':exit' Leave chat room\n- - - - - - - - - -\n")
}

func showConversation() {
	printHeader()
	for _, m := range messages {
		if m.Type == "text" {
			fmt.Println(m.User + ": " + m.Body)
		} else {
			fmt.Println(m.User + ": file[" + m.Body + "]")
		}
	}
	fmt.Print("\n > ")
}

func client(c net.Conn) {
	var rec message.Message
	for {
		err := gob.NewDecoder(c).Decode(&rec)
		if err != nil {
			return
		}
		messages = append(messages, rec)
		go showConversation()
	}
}

func sendMessage(c net.Conn, _user, _body, _type string, _name []byte) {
	m := message.Message{
		User: _user,
		Body: _body,
		Type: _type,
		File: _name,
	}
	err := gob.NewEncoder(c).Encode(m)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	c, err := net.Dial("tcp", "192.168.100.9:5400")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer c.Close()

	var input string
	var user string
	s := bufio.NewScanner(os.Stdin)

	fmt.Print("User: ")
	fmt.Scanln(&user)
	gob.NewEncoder(c).Encode(user)

	go client(c)

	printHeader()

	for input != ":exit" {
		fmt.Print(" > ")
		s.Scan()
		input = s.Text()
		if input[0] == ':' {
			if len(input) > 4 && input[1:5] == "exit" {
				fmt.Println("Disconnected")
				return
			} else if len(input) > 6 && input[1:5] == "file" {
				file, err := ioutil.ReadFile(input[6:])
				if err != nil {
					fmt.Println(err)
				} else {
					sendMessage(c, user, input[6:], "file", file)
				}
			} else if len(input) > 10 && input[1:9] == "download" {
				for _, m := range messages {
					if m.Body == input[10:] {
						err := ioutil.WriteFile(m.Body, m.File, 0644)
						if err != nil {
							fmt.Println("Unable to download the file:\n", err)
						}
						break
					}
				}
			} else {
				fmt.Println("Command unavailable(", input, ")")
			}
		} else {
			sendMessage(c, user, input, "text", []byte("none"))
		}
	}
}
