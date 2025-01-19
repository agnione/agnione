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

#########################################################################################
*/
package wshandler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second       // Time allowed to write a message to the peer.
	readWait       = 500 * time.Millisecond // Time allowed to read the next pong message from the peer.
	maxMessageSize = 512                    // Maximum message size allowed from peer.
)

type MonitorLevel int8

const (
	STAUS_MONITOR    MonitorLevel = 0
	LOG_MONITOR      MonitorLevel = 1
	ACTIVITY_MONITOR MonitorLevel = 2
)

var newline = []byte{'\n'}

// / websocket upgrade options
var upgrader = websocket.Upgrader{
	ReadBufferSize:    1024,
	WriteBufferSize:   1024,
	EnableCompression: true,
	CheckOrigin: func(pRequest *http.Request) bool {
        return true
    },
}

// Client is a middleman between the websocket connection and the pWSHub.
type WSClient struct {
	pWSHub          *WSHub          /// ws control pWSHub instance
	conn         *websocket.Conn // The websocket connection.
	send         chan []byte     // Buffered channel of outbound messages.
	Monitor_Type MonitorLevel
}

// reader pumps messages from the websocket connection to the pWSHub.
//
// The application runs reader in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *WSClient) reader() {

	_ticker := time.NewTicker(readWait) /// ticker to read in intervals

	defer func() {
		if _r:=recover();_r!=nil{
			fmt.Println("Recovered panic ",_r)
			_r=nil
		}
		
		/// clear and exit
		_ticker.Stop()
		_ticker = nil
		c.pWSHub.unregister <- c /// set falg to unregister from pWSHub
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	
	for {
		select {
		case <-c.pWSHub.stopped: /// if pWSHub stopped
			return
		case <-_ticker.C: /// time to read
			_, _, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err,  websocket.CloseAbnormalClosure,websocket.CloseInternalServerErr) {
					log.Printf("error: %v", err)
				}
				return
			}
		}
	}
}

// writer pumps messages from the pWSHub to the websocket connection.
//
// A goroutine running writer is started for each connection.
// The application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (wsc *WSClient) writer() {
	
	/// declare variables
	var _writer io.WriteCloser
	var _err error
	var _message []byte
	var _isOK bool
	var _msgCount = 0

	defer func() {
		if _r:=recover();_r!=nil{
			fmt.Println("Recovered panic ",_r)
			
		}
		
		/// clear and exit
		_writer = nil
		_message = nil
		_err = nil
		_msgCount = 0
		
	}()

	defer func ()  {
		wsc.conn.Close()
	}()
	
	var _index int
	
	for {

		select {
		case <-wsc.pWSHub.stopped: /// if pWSHub stopped
			return
		case _message, _isOK = <-wsc.send: /// if we have a message to send
			wsc.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !_isOK { // The pWSHub closed the channel.
				wsc.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			_writer, _err = wsc.conn.NextWriter(websocket.TextMessage)
			if _err != nil { /// failed to crate the writer
				return
			}

			_writer.Write(_message) /// writes the message

			// Incase, Add queued messages to the current websocket message.
			_msgCount = len(wsc.send)
			for _index= 0; _index < _msgCount; _index++ {
				_writer.Write(slices.Concat(newline,<-wsc.send))
			}

			if _err = _writer.Close(); _err != nil { /// try to close the writer
				return
			}

			/// clear currnet values
			_writer = nil
			_message = nil
		}

	}
}

func serve_client(pWSHub *WSHub, pResWriter http.ResponseWriter, pRequest *http.Request, pClient_Type MonitorLevel) {

	/// upgrade the http connection to web socket	
	_conn, _err := upgrader.Upgrade(pResWriter, pRequest, nil)
	if _err != nil {
		log.Println(fmt.Errorf("failed to upgrade http to web socket %v", _err)) /// failed to upgrage to http to ws
		return
	}

	///  creats a client instance
	_client := &WSClient{pWSHub: pWSHub, conn: _conn, send: make(chan []byte, 256), Monitor_Type: pClient_Type}
	_client.pWSHub.register <- _client /// set the client to pWSHub

	go _client.writer() /// start the mesage writer
	go _client.reader() /// starts the message reader

}

// serveWs handles websocket requests from the peers.
func ServeWs(pWSHub *WSHub, pResWriter http.ResponseWriter, pRequest *http.Request) {
	serve_client(pWSHub, pResWriter, pRequest, ACTIVITY_MONITOR)
}

// serveWs handles websocket requests from the peers.
func ServeWSStatus(pWSHub *WSHub, pResWriter http.ResponseWriter, pRequest *http.Request) {
	serve_client(pWSHub, pResWriter, pRequest, STAUS_MONITOR)
}

// serveWs handles websocket requests from the peers.
func ServeWSLogTailer(pWSHub *WSHub, pResWriter http.ResponseWriter, pRequest *http.Request) {
	serve_client(pWSHub, pResWriter, pRequest, LOG_MONITOR)
}
