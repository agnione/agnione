//#########################################################################################
// Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 25/01/2024
//
// Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
//						Licensed under the Apache License, Version 2.0 (the "License");
//						you may not use this file except in compliance with the License.
//						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
//
//						Unless required by applicable law or agreed to in writing, software
//						distributed under the License is distributed on an "AS IS" BASIS,
//						WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//						See the License for the specific language governing permissions and
//						limitations under the License.
//
// Class/module  :   WsHub
//
// Objective     :   Define the centerlize hub for web socket client management
//					This package will be used to mamange the web socket clients and broadcase message
//					among clients.
//#########################################################################################
// Author                        Date        Action      Description
//------------------------------------------------------------------------------------------------------------
// Ajith de Silva				25/01/2024	Created 	Created the initial version
// Ajith de Silva				25/01/2024	Addes    	Added function to enable boradcasting message
// Ajith de Silva				29/01/2024	Added 		Added function to return the web socket client count
// Ajith de Silva				01/02/2024	Added 		Added startus monitor related features
//#########################################################################################

package ssehandler

import "log"

const MAX_CLIENTS = 50 //// define MAX clients to 20 per topic

// Hub maintains the set of active clients and broadcasts messages to the clients.
type SSEHub struct {
	lclients map[*SSEClient]bool // log reader clients.
	mclients map[*SSEClient]bool // monitoring clients.
	sclients map[*SSEClient]bool // status clients.

	broadcast_monitor chan *SSE_Event // outboud message to boradcast
	broadcast_status  chan *SSE_Event // outboud status message to boradcast
	broadcast_log     chan *SSE_Event // outboud log entries to boradcast
	register          chan *SSEClient // Register requests from the clients.
	unregister        chan *SSEClient // Unregister requests from clients.

	logger              *log.Logger
	stopped             chan bool
	event_client_count  uint8
	status_client_count uint8
	log_client_count    uint8
}

func NewHub(pLogger *log.Logger) *SSEHub {
	/// create a new instace of the WSHub
	return &SSEHub{
		broadcast_monitor: make(chan *SSE_Event),
		broadcast_status:  make(chan *SSE_Event),
		broadcast_log:     make(chan *SSE_Event),

		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),

		mclients:            make(map[*SSEClient]bool),
		sclients:            make(map[*SSEClient]bool),
		lclients:            make(map[*SSEClient]bool),
		stopped:             make(chan bool),
		logger:              pLogger,
		event_client_count:  0,
		status_client_count: 0,
		log_client_count:    0,
	}
}

func (h *SSEHub) DeInitialize() {
	h.broadcast_log = nil
	h.broadcast_monitor = nil
	h.broadcast_status = nil
	h.clear_clients()
	h.register = nil
	h.unregister = nil
	h.stopped = nil
}

func (h *SSEHub) clear_clients() {
	for s := range h.mclients {
		delete(h.mclients, s)
	}
	h.mclients = nil

	for s := range h.lclients {
		delete(h.lclients, s)
	}
	h.lclients = nil

	for s := range h.sclients {
		delete(h.sclients, s)
	}
	h.sclients = nil

	h.event_client_count = 0
	h.status_client_count = 0
	h.log_client_count = 0

}

// Run executes the main functionality of the Hub.
// It manages the new client registrations, client unregistrations.
// Also breadcasting messages among web scokcet clients
func (h *SSEHub) Run() {

	var _client *SSEClient
	var _message *SSE_Event

	defer func() {
		_client = nil

	}()

	for {
		select {

		case <-h.stopped: /// if stop requested
			h.clear_clients()

			return

		case _client = <-h.register: /// if new client connects
			{
				switch _client.Monitor_Type {

				case ACTIVITY_MONITOR:
					if MAX_CLIENTS > h.event_client_count+1 {
						h.mclients[_client] = true
						h.event_client_count++
					} else {
						_client = nil
					}
				case STAUS_MONITOR:
					if MAX_CLIENTS > h.status_client_count+1 {
						h.sclients[_client] = true
						h.status_client_count++
					} else {
						_client = nil
					}
				case LOG_MONITOR:
					if MAX_CLIENTS > h.log_client_count+1 {
						h.lclients[_client] = true
						h.log_client_count++
					} else {
						_client = nil
					}
				}
			}
			_client = nil

		case _client = <-h.unregister: /// if client disconnects

			/// do the decrent of client count based on the client type
			switch _client.Monitor_Type {
			case ACTIVITY_MONITOR:
				if _, _ok := h.mclients[_client]; _ok {
					h.mclients[_client] = false
					delete(h.mclients, _client)
					close(_client.event_message)
					h.event_client_count--
				}

			case STAUS_MONITOR:
				if _, _ok := h.sclients[_client]; _ok {
					h.sclients[_client] = false
					delete(h.sclients, _client)
					close(_client.event_message)
					h.status_client_count--
				}

			case LOG_MONITOR:
				if _, _ok := h.lclients[_client]; _ok {
					h.lclients[_client] = false
					delete(h.lclients, _client)
					close(_client.event_message)
					h.log_client_count--
				}

			}

		case _message = <-h.broadcast_monitor: /// if broadcast message is requested

			for __client := range h.mclients {
				__client.event_message <- _message
			}

		case _message = <-h.broadcast_status: /// if broadcast status is requested
			for __client := range h.sclients {
				__client.event_message <- _message
			}

		case _message = <-h.broadcast_log: /// if broadcast log entries is requested

			for __client := range h.lclients {
				__client.event_message <- _message
			}
		}
	}
}

// BroadCast boradcasts message among connected websocket clients
func (h *SSEHub) Broadcast_Event(pMessage SSE_Event) {
	h.broadcast_monitor <- &pMessage
}

// BroadCast boradcast status message among connected websocket clients
func (h *SSEHub) Broadcast_Status(pMessage SSE_Event) {
	h.broadcast_status <- &pMessage
}

// BroadCast boradcast log entries among connected websocket clients
func (h *SSEHub) Broadcast_Log(pMessage SSE_Event) {
	h.broadcast_log <- &pMessage
}

// ClientsCount returns the connected web socket client count for monitor end point
func (h *SSEHub) Event_Clients_Count() uint8 {
	return h.event_client_count
}

// StatusClientsCount returns the connected web socket client count for status endpoint
func (h *SSEHub) Status_Clients_Count() uint8 {
	return h.status_client_count

}

// LogClientsCount returns the connected web socket client count for log endpoint
func (h *SSEHub) Log_Clients_Count() uint8 {
	return h.log_client_count

}

// Stop stops the hub.
func (h *SSEHub) Stop() {
	h.stopped <- true
}
