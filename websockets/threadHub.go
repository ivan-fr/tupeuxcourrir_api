// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package websockets

import "tupeuxcourrir_api/models"

var mapThreadHub = make(map[int]*ThreadHub)

// ThreadHub maintains the set of active Clients and broadcasts messages to the
// Clients.
type ThreadHub struct {
	threadID int

	// Registered Clients.
	Clients map[*ThreadClient]bool

	// Inbound messages from the Clients.
	Broadcast chan *models.Message

	// Register requests from the Clients.
	Register chan *ThreadClient

	// Unregister requests from Clients.
	Unregister chan *ThreadClient
}

func GetThreadHub(thread *models.Thread) *ThreadHub {
	threadHub, ok := mapThreadHub[thread.IdThread]

	if !ok {
		mapThreadHub[thread.IdThread] = &ThreadHub{
			threadID:   thread.IdThread,
			Broadcast:  make(chan *models.Message),
			Register:   make(chan *ThreadClient),
			Unregister: make(chan *ThreadClient),
			Clients:    make(map[*ThreadClient]bool),
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
		case client := <-tH.Register:
			tH.Clients[client] = true
		case client := <-tH.Unregister:
			if _, ok := tH.Clients[client]; ok {
				delete(tH.Clients, client)
				close(client.Send)
				if len(tH.Clients) == 0 {
					stopLoop = true
				}
			}
		case message := <-tH.Broadcast:
			for client := range tH.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(tH.Clients, client)
				}
			}
		}

		if stopLoop {
			break
		}
	}

	delete(mapThreadHub, tH.threadID)
}
