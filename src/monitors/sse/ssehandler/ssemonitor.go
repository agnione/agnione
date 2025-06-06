package ssehandler

import (
	"log"
	"net/http"
)

// SSEMonitor structure
type SSEMonitor struct {
	sseHub      *SSEHub
	messge      chan string /// message channel
	logger      *log.Logger
	isstarted   bool /// flag to set the start/stop status
	initialized bool /// flag to set the start/stop status
}

// Initialize initializes the Initialize instance with default values.
// websocket.Upgrader with ReadBufferSize:1024,	WriteBufferSize:1024 and EnableCompression:true
func (ssem *SSEMonitor) Initialize(pLogger *log.Logger) bool {

	/// initialize the web socket upgrader
	ssem.logger = pLogger

	ssem.sseHub = NewHub(pLogger)   /// create an instance of websocket hub
	ssem.messge = make(chan string) /// crete the message channel
	ssem.isstarted = false
	ssem.initialized = true
	return ssem.initialized
}

// DeInitialize clear the data and instance of objects.
func (ssem *SSEMonitor) DeInitialize() {
	ssem.messge = nil
	ssem.sseHub = nil
}

func (ssem *SSEMonitor) Start() bool {

	go ssem.sseHub.Run()
	ssem.isstarted = true
	return ssem.isstarted
}

func (ssem *SSEMonitor) Stop() bool {

	ssem.sseHub.Stop()
	ssem.isstarted = false
	return true
}

func (ssem *SSEMonitor) IsStarted() bool {
	return ssem.isstarted
}

// BroadCast broadcasts a message among to the clients who connected to the monitor endpoint
func (ssem *SSEMonitor) Broadcast_Event(pMessage SSE_Event) {
	defer recover()
	ssem.sseHub.Broadcast_Event(pMessage)
}

// BroadCastStatus broadcasts message to the clients who connected to the status endpoint
func (ssem *SSEMonitor) Broadcast_Status(pMessage SSE_Event) {
	defer recover()
	ssem.sseHub.Broadcast_Status(pMessage)
}

// BroadCast broadcasts a message among to the clients connected to the monitor endpoint
func (ssem *SSEMonitor) Broadcast_Log(pMessage SSE_Event) {
	defer recover()
	ssem.sseHub.Broadcast_Log(pMessage)
}

// / MonitorClientsCount returns the number of clients connected to monitor endpoint
func (ssem *SSEMonitor) MonitorClientsCount() uint8 {
	if ssem.sseHub != nil {
		return ssem.sseHub.Event_Clients_Count()
	} else {
		return 0
	}

}

// / MonitorClientsCount returns the number of clients connected to status endpoint
func (ssem *SSEMonitor) StatusClientsCount() uint8 {
	if ssem.sseHub != nil {
		return ssem.sseHub.Status_Clients_Count()
	} else {
		return 0
	}
}

// / MonitorClientsCount returns the number of clients connected to status endpoint
func (ssem *SSEMonitor) LogClientsCount() uint8 {
	if ssem.sseHub != nil {
		return ssem.sseHub.Log_Clients_Count()
	} else {
		return 0
	}
}

// / set the events endpoint
func (ssem *SSEMonitor) Monitor(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	Serve_Monitor(ssem.sseHub, w, r)
}

// / set the logger endpoint
func (ssem *SSEMonitor) Logger(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	Serve_Logger(ssem.sseHub, w, r)
}

// / set the status endpoint
func (ssem *SSEMonitor) Status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	Serve_Status(ssem.sseHub, w, r)
}
