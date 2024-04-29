package main

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var addr = flag.String("addr", "localhost:8089", "http service address")

func LoadBRC20InputData(fname string) ([][]byte, error) {
	var contentMap map[string]bool = make(map[string]bool, 0)

	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	max := 4 * 1024 * 1024
	buf := make([]byte, max)
	scanner.Buffer(buf, max)

	var brc20Scriptions [][]byte
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Split(line, " ")

		if len(fields) != 12 {
			return nil, fmt.Errorf("invalid data format")
		}

		if _, ok := contentMap[fields[7]]; ok {
			continue
		} else {
			content, err := hex.DecodeString(fields[7])
			if err != nil {
				return nil, err
			}
			contentMap[fields[7]] = true
			brc20Scriptions = append(brc20Scriptions, content)

		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return brc20Scriptions, nil
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	sources, err := LoadBRC20InputData("./data/brc20.input.txt")
	if err != nil {
		log.Fatal("dial:", err)
	}
	index := 0

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(sources[index]))
			if err != nil {
				log.Println("write:", err)
				return
			}
			index += 1
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
