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

package wshandler

const MAX_CLIENTS = 20	//// define MAX WS clients to 20

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type WSHub struct {
	clients              map[*WSClient]bool // Registered clients.
	broadcast_monitor    chan []byte        // outboud message to boradcast
	broadcast_status     chan []byte        // outboud status message to boradcast
	broadcast_log        chan []byte        // outboud log entries to boradcast
	register             chan *WSClient     // Register requests from the clients.
	unregister           chan *WSClient     // Unregister requests from clients.
	stopped              chan bool
	monitor_client_count uint8
	status_client_count  uint8
	log_client_count     uint8
	
}

func NewHub() *WSHub {
	/// create a new instace of the WSHub
	return &WSHub{
		broadcast_monitor:    make(chan []byte),
		broadcast_status:     make(chan []byte),
		broadcast_log:        make(chan []byte),
		register:             make(chan *WSClient),
		unregister:           make(chan *WSClient),
		clients:              make(map[*WSClient]bool),
		stopped:              make(chan bool),
		monitor_client_count: 0,
		status_client_count:  0,
		log_client_count:     0,
	}
}

// Run executes the main functionality of the Hub.
// It manages the new client registrations, client unregistrations.
// Also breadcasting messages among web scokcet clients
func (h *WSHub) Run() {
	
	var _client *WSClient
	var _message []byte
	
	defer func(){
		_client=nil
		_message=nil
	}()
	
	for {
		select {
			
		case <-h.stopped: /// if stop requested
			for _client = range h.clients {
				delete(h.clients, _client)
				close(_client.send)
			}
			_client=nil
			h.monitor_client_count = 0
			h.status_client_count = 0
			h.log_client_count = 0
			return

		case _client = <-h.register: /// if new client connects
			{
				switch _client.Monitor_Type {

				case ACTIVITY_MONITOR:
					if MAX_CLIENTS > h.monitor_client_count+1 {
						h.clients[_client] = true
						h.monitor_client_count++
					} else {
						_client = nil
					}
				case STAUS_MONITOR:
					if MAX_CLIENTS > h.status_client_count+1 {
						h.clients[_client] = true
						h.status_client_count++
					} else {
						_client = nil
					}
				case LOG_MONITOR:
					if MAX_CLIENTS > h.log_client_count+1 {
						h.clients[_client] = true
						h.log_client_count++
					} else {
						_client = nil
					}
				}
			}
			_client=nil
			
		case _client = <-h.unregister: /// if client disconnects
			if _, _ok := h.clients[_client]; _ok {
				h.clients[_client] = false
				delete(h.clients, _client)
				close(_client.send)

				/// do the decrent of client count based on the client type
				switch _client.Monitor_Type {

					case ACTIVITY_MONITOR:
						h.monitor_client_count--

					case STAUS_MONITOR:
						h.status_client_count--

					case LOG_MONITOR:
						h.log_client_count--

				}
			}
			_client=nil
			
		case _message = <-h.broadcast_monitor: /// if broadcast message is requested
			for _client = range h.clients {
				if _client.Monitor_Type == ACTIVITY_MONITOR {
					select {
					case _client.send <- _message:
					default:
						close(_client.send)
						delete(h.clients, _client)
					}
				}
			}
			_client=nil
			
		case _message = <-h.broadcast_status: /// if broadcast status is requested
			for _client = range h.clients {
				if _client.Monitor_Type == STAUS_MONITOR {
					select {
					case _client.send <- _message:
					default:
						close(_client.send)
						delete(h.clients, _client)
					}
				}
				_client=nil

			}
		case _message = <-h.broadcast_log: /// if broadcast log entries is requested
			for _client = range h.clients {
				if _client.Monitor_Type == LOG_MONITOR {
					select {
						case _client.send <- _message:
						default:
							close(_client.send)
							delete(h.clients, _client)
					}
				}
			}
			_client=nil
		}
	}
}

// BroadCast boradcasts message among connected websocket clients
func (h *WSHub) BroadCast(pMessage []byte) {
	h.broadcast_monitor <- pMessage
}

// BroadCast boradcast status message among connected websocket clients
func (h *WSHub) BroadCastStatus(pMessage []byte) {
	h.broadcast_status <- pMessage
}

// BroadCast boradcast log entries among connected websocket clients
func (h *WSHub) BroadCastLogEntries(pMessage []byte) {
	h.broadcast_log <- pMessage
}

// ClientsCount returns the connected web socket client count for monitor end point
func (h *WSHub) MonitorClientsCount() uint8 {
	if h.clients != nil {
		return h.monitor_client_count
	} else {
		return 0
	}
}

// StatusClientsCount returns the connected web socket client count for status endpoint
func (h *WSHub) StatusClientsCount() uint8 {
	if h.clients != nil {
		return h.status_client_count
	} else {
		return 0
	}
}

// LogClientsCount returns the connected web socket client count for log endpoint
func (h *WSHub) LogClientsCount() uint8 {
	if h.clients != nil {
		return h.log_client_count
	} else {
		return 0
	}
}

// Stop stops the hub.
func (h *WSHub) Stop() {
	h.stopped <- true
}
