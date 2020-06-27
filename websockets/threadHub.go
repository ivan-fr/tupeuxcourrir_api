// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websockets

import "tupeuxcourrir_api/models"

var mapThreadHub = make(map[int]*ThreadHub)

// ThreadHub maintains the set of active clients and broadcasts messages to the
// clients.
type ThreadHub struct {
	threadID int

	// Registered clients.
	clients map[*ThreadClient]bool

	// Inbound messages from the clients.
	broadcast chan *models.Message

	// Register requests from the clients.
	register chan *ThreadClient

	// Unregister requests from clients.
	unregister chan *ThreadClient
}

func GetThreadHub(thread *models.Thread) *ThreadHub {
	threadHub, ok := mapThreadHub[thread.IdThread]

	if !ok {
		mapThreadHub[thread.IdThread] = &ThreadHub{
			threadID:   thread.IdThread,
			broadcast:  make(chan *models.Message),
			register:   make(chan *ThreadClient),
			unregister: make(chan *ThreadClient),
			clients:    make(map[*ThreadClient]bool),
		}

		defer func() {
			go mapThreadHub[thread.IdThread].run()
		}()

		return mapThreadHub[thread.IdThread]
	}

	return threadHub
}

func (tH *ThreadHub) run() {
	stopLoop := false

	for {
		select {
		case client := <-tH.register:
			tH.clients[client] = true
		case client := <-tH.unregister:
			if _, ok := tH.clients[client]; ok {
				delete(tH.clients, client)
				close(client.send)
				if len(tH.clients) == 0 {
					stopLoop = true
				}
			}
		case message := <-tH.broadcast:
			for client := range tH.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(tH.clients, client)
				}
			}
		}

		if stopLoop {
			break
		}
	}

	delete(mapThreadHub, tH.threadID)
}
