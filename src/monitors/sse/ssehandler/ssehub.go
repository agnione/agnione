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
// Ajith de Silva				01/02/2024	Added 		Added status monitor related features
// Ajith de Silva				06/06/2025	Updated 	Updated the package to use sync.Map for client management
//#########################################################################################

package ssehandler

import (
	"log"
	"sync"
	"sync/atomic"
)

const MAX_CLIENTS = 50 //// define MAX clients to 20 per topic

// Hub maintains the set of active clients and broadcasts messages to the clients.
type SSEHub struct {
	lclients sync.Map // log reader clients.
	mclients sync.Map // monitoring clients.
	sclients sync.Map // status clients.

	broadcast_monitor chan *SSE_Event // outboud message to boradcast
	broadcast_status  chan *SSE_Event // outboud status message to boradcast
	broadcast_log     chan *SSE_Event // outboud log entries to boradcast
	register          chan *SSEClient // Register requests from the clients.
	unregister        chan *SSEClient // Unregister requests from clients.

	logger              *log.Logger
	stopped             chan bool
	event_client_count  atomic.Int32
	status_client_count atomic.Int32
	log_client_count    atomic.Int32
}

func NewHub(pLogger *log.Logger) *SSEHub {
	/// create a new instace of the WSHub
	return &SSEHub{
		broadcast_monitor: make(chan *SSE_Event),
		broadcast_status:  make(chan *SSE_Event),
		broadcast_log:     make(chan *SSE_Event),

		register:   make(chan *SSEClient),
		unregister: make(chan *SSEClient),

		mclients: sync.Map{}, // activity clients
		sclients: sync.Map{}, // status clients
		lclients: sync.Map{}, // log clients
		stopped:  make(chan bool),
		logger:   pLogger,
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

	h.lclients.Range(func(key any, value any) bool {
		h.mclients.Delete(key)
		return true
	})

	h.sclients.Range(func(key any, value any) bool {
		h.mclients.Delete(key)
		return true
	})

	h.event_client_count.Store(0)
	h.status_client_count.Store(0)
	h.log_client_count.Store(0)

}

// Run executes the main functionality of the Hub.
// It manages the new client registrations, client unregistrations.
// Also breadcasting messages among web socket clients
func (h *SSEHub) Run() {

	var _client *SSEClient
	var _message *SSE_Event

	defer func() {
		_client = nil

	}()

	h.event_client_count.Store(0)
	h.status_client_count.Store(0)
	h.log_client_count.Store(0)

	for {
		select {

		case <-h.stopped: /// if stop requested
			h.clear_clients()
			return

		case _client = <-h.register: /// if new client connects
			h.register_client(_client)

		case _client = <-h.unregister: /// if client disconnects

			/// do the decrent of client count based on the client type
			h.unregistre_client(_client)

		case _message = <-h.broadcast_monitor: /// if broadcast message is requested

			h.mclients.Range(func(key any, value any) bool {
				key.(*SSEClient).event_message <- _message
				return true
			})

		case _message = <-h.broadcast_status: /// if broadcast status is requested
			h.sclients.Range(func(key any, value any) bool {
				key.(*SSEClient).event_message <- _message
				return true
			})

		case _message = <-h.broadcast_log: /// if broadcast log entries is requested

			h.lclients.Range(func(key any, value any) bool {
				key.(*SSEClient).event_message <- _message
				return true
			})
		}
	}
}

func (h *SSEHub) unregistre_client(_client *SSEClient) {

	defer recover()
	switch _client.Monitor_Type {
	case ACTIVITY_MONITOR:
		if _, _ok := h.mclients.Load(_client); _ok {
			h.mclients.Delete(_client)
			close(_client.event_message)
			h.event_client_count.Add(^int32(0))
		}

	case STAUS_MONITOR:
		if _, _ok := h.sclients.Load(_client); _ok {
			h.sclients.Delete(_client)
			close(_client.event_message)
			h.status_client_count.Add(^int32(0))
		}

	case LOG_MONITOR:
		if _, _ok := h.lclients.Load(_client); _ok {
			h.lclients.Delete(_client)
			close(_client.event_message)
			h.log_client_count.Add(^int32(0))
		}

	}
}

func (h *SSEHub) register_client(_client *SSEClient) {

	defer recover()

	switch _client.Monitor_Type {

	case ACTIVITY_MONITOR:
		if MAX_CLIENTS > h.event_client_count.Load()+1 {
			h.mclients.Store(_client, true)
			h.event_client_count.Add(1)
		}
	case STAUS_MONITOR:
		if MAX_CLIENTS > h.status_client_count.Load()+1 {
			h.sclients.Store(_client, true)
			h.status_client_count.Add(1)
		}
	case LOG_MONITOR:
		if MAX_CLIENTS > h.log_client_count.Load()+1 {
			h.lclients.Store(_client, true)
			h.log_client_count.Add(1)
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
	return uint8(h.event_client_count.Load())
}

// StatusClientsCount returns the connected web socket client count for status endpoint
func (h *SSEHub) Status_Clients_Count() uint8 {
	return uint8(h.status_client_count.Load())

}

// LogClientsCount returns the connected web socket client count for log endpoint
func (h *SSEHub) Log_Clients_Count() uint8 {
	return uint8(h.log_client_count.Load())

}

// Stop stops the hub.
func (h *SSEHub) Stop() {
	h.stopped <- true
}
