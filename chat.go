package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
)

type Client struct {
	conn     net.Conn
	nickname string
	ch       chan Message
}

type Message struct {
	from    string
	to      string
	content string
}

func main() {
	ln, err := net.Listen("tcp", ":6000")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	msgchan := make(chan Message)
	addchan := make(chan Client)
	rmchan := make(chan Client)

	go handleMessages(msgchan, addchan, rmchan)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println(err)
			continue
		}

		go handleConnection(conn, msgchan, addchan, rmchan)
	}
}

func sendAuth(c net.Conn, bufc *bufio.Reader) string {
	io.WriteString(c, "auth")
	nick, _, _ := bufc.ReadLine()
	return string(nick)
}

func (c Client) Receive(ch chan<- Message) {
	bufc := bufio.NewReader(c.conn)
	for {
		line, err := bufc.ReadString('\n')
		if err != nil {
			break
		}
		tmp := strings.Split(line, ":")
		message := Message{
			from:    c.nickname,
			to:      tmp[0],
			content: fmt.Sprintf("%s: %s", c.nickname, tmp[1]),
		}

		ch <- message
	}
}

func (c Client) Send(ch <-chan Message) {
	for msg := range ch {
		_, err := io.WriteString(c.conn, msg.content)
		if err != nil {
			return
		}
	}
}

func handleConnection(c net.Conn, msgchan chan<- Message, addchan chan<- Client, rmchan chan<- Client) {
	bufc := bufio.NewReader(c)
	defer c.Close()
	client := Client{
		conn:     c,
		nickname: sendAuth(c, bufc),
		ch:       make(chan Message),
	}
	if strings.TrimSpace(client.nickname) == "" {
		io.WriteString(c, "Invalid Username\n")
		return
	}

	// Register user
	addchan <- client
	defer func() {
		log.Printf("Connection from %v closed.\n", c.RemoteAddr())
		rmchan <- client
	}()
	io.WriteString(c, fmt.Sprintf("Welcome, %s!\n\n", client.nickname))

	// I/O
	go client.Receive(msgchan)
	client.Send(client.ch)
}

func handleMessages(msgchan <-chan Message, addchan <-chan Client, rmchan <-chan Client) {
	clients := make(map[string]Client)

	for {
		select {
		case msg := <-msgchan:
			log.Printf("New message: %s", msg)
			for _, client := range clients {
				if client.nickname == msg.to {
					go func(mch chan<- Message) {
						mch <- msg
					}(client.ch)
				}
			}
		case client := <-addchan:
			log.Printf("New client: %v\n", client.conn)
			clients[client.nickname] = client
		case client := <-rmchan:
			log.Printf("Client disconnects: %v\n", client)
			delete(clients, client.nickname)
		}
	}
}
