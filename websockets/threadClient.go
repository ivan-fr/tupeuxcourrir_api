// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websockets

import (
	"github.com/gorilla/websocket"
	"log"
	"time"
	"tupeuxcourrir_api/config"
	"tupeuxcourrir_api/models"
)

// ThreadClient is a middleman between the websocket connection and the ThreadHub.
type ThreadClient struct {
	IdUser int

	ThreadHub *ThreadHub

	// The websocket connection.
	Conn *websocket.Conn

	// Buffered channel of outbound messages.
	Send chan *models.Message
}

// ReadPump pumps messages from the websocket connection to the ThreadHub.
//
// The application runs ReadPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *ThreadClient) ReadPump() {
	defer func() {
		c.ThreadHub.Unregister <- c
		_ = c.Conn.Close()
	}()

	c.Conn.SetReadLimit(config.MaxMessageSize)
	_ = c.Conn.SetReadDeadline(time.Now().Add(config.PongWait))
	c.Conn.SetPongHandler(func(string) error {
		_ = c.Conn.SetReadDeadline(time.Now().Add(config.PongWait))
		return nil
	})

	for {
		var message models.Message

		err := c.Conn.ReadJSON(&message)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.ThreadHub.Broadcast <- &message
	}
}

// WritePump pumps messages from the ThreadHub to the websocket connection.
//
// A goroutine running WritePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *ThreadClient) WritePump() {
	ticker := time.NewTicker(config.PingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if !ok {
				// The ThreadHub closed the channel.
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteJSON(message)

			if err != nil {
				return
			}
		case <-ticker.C:
			_ = c.Conn.SetWriteDeadline(time.Now().Add(config.WriteWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
