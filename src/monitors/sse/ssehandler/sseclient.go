// wsclient implements read and writes data from the connected websocket clients
//
// This package defined functions:
//	- Initialize
//	- DeInitialize
//	- Start
//	- Stop
//	- BroadCast
//	- BroadCastStatus
//	- MonitorClientsCount
//	- StatusClientsCount
//	- NewHub
//	- Run
/*
#########################################################################################

	Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 24/01/2024
	Class/module  :   WSClient
	Objective     :   Define the package for support web socket client connections
	This package has functions to read and write features to connected client
#########################################################################################
	Author                 	Date        	Action      	Description
-----------------------------------------------------------------------------------------------------------------
	Ajith de Silva		24/01/2024	Created 	Created the initial version
	Ajith de Silva		29/01/2024	Updated 	Defined functions with parameters & return values
	Ajith de Silva		29/01/2024	Updated 	Implemented functions
	Ajith de Silva		01/02/2024	Added 		Added the status endpoint to monitor status vis web socket
	Ajith de Silva		06/06/2025	Updated 	Updated the package to use sync.Map for client management
#########################################################################################
*/
package ssehandler

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type MonitorLevel int8

const (
	STAUS_MONITOR    MonitorLevel = 0
	LOG_MONITOR      MonitorLevel = 1
	ACTIVITY_MONITOR MonitorLevel = 2
)

type SSE_Event struct {
	ID      string
	Message string
}

// Client is a middleman between the websocket connection and the pWSHub.
type SSEClient struct {
	pWSHub         *SSEHub                  /// ws control pWSHub instance
	writer         *http.ResponseWriter     // The websocket connection.
	res_controller *http.ResponseController // The websocket connection.
	context        context.Context
	event_message  chan *SSE_Event // Buffered channel of outbound messages.
	Monitor_Type   MonitorLevel
	envnt_name     string
}

// writer pumps messages from the pWSHub to the websocket connection.
//
// A goroutine running writer is started for each connection.
// The application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (wsc *SSEClient) Event_Writer() {

	/// declare variables

	var _err error
	var _ev_msg *SSE_Event
	var _isOK bool

	defer func() {
		if _r := recover(); _r != nil {
			fmt.Println("Recovered panic ", _r)
		}

		/// clear and exit
		_ev_msg = nil
		_err = nil
	}()

	for {

		select {
		case <-wsc.pWSHub.stopped: /// if pWSHub stopped
			wsc.pWSHub.logger.Println("SSE STOPPED")
			return
		case <-wsc.context.Done():
			wsc.pWSHub.logger.Println("Client Gone")
			return

		case _ev_msg, _isOK = <-wsc.event_message: /// if we have a message to send

			if !_isOK {
				return
			}

			_, _err = fmt.Fprintf(*wsc.writer, "id: %s\nevent: %s\ndata: %s %s\n\n", _ev_msg.ID, wsc.envnt_name, time.Now().Format("2006-01-02 15:04:05"), _ev_msg.Message)
			if _err != nil {
				return
			}
			_err = wsc.res_controller.Flush()
			if _err != nil {
				return
			}
		}

	}
}

func serve_client(pWSHub *SSEHub, pResWriter http.ResponseWriter, pRequest *http.Request, pClient_Type MonitorLevel, pEventName string) {

	defer recover()

	///  creats a client instance
	_client := &SSEClient{pWSHub: pWSHub, writer: &pResWriter, event_message: make(chan *SSE_Event, 1), envnt_name: pEventName,
		Monitor_Type: pClient_Type, res_controller: http.NewResponseController(pResWriter), context: pRequest.Context()}

	defer func() {
		_client = nil
	}()

	pWSHub.register <- _client

	go _client.Event_Writer()

	<-pRequest.Context().Done()

	pWSHub.unregister <- _client
}

// serveWs handles websocket requests from the peers.
func Serve_Monitor(pWSHub *SSEHub, pResWriter http.ResponseWriter, pRequest *http.Request) {
	serve_client(pWSHub, pResWriter, pRequest, ACTIVITY_MONITOR, "events")
}

// serveWs handles websocket requests from the peers.
func Serve_Status(pWSHub *SSEHub, pResWriter http.ResponseWriter, pRequest *http.Request) {
	serve_client(pWSHub, pResWriter, pRequest, STAUS_MONITOR, "status")
}

// serveWs handles websocket requests from the peers.
func Serve_Logger(pWSHub *SSEHub, pResWriter http.ResponseWriter, pRequest *http.Request) {
	serve_client(pWSHub, pResWriter, pRequest, LOG_MONITOR, "logs")
}

// serveWs handles websocket requests from the peers.
func Serve_All(pWSHub *SSEHub, pResWriter http.ResponseWriter, pRequest *http.Request) {
	serve_client(pWSHub, pResWriter, pRequest, LOG_MONITOR, "*")
}
