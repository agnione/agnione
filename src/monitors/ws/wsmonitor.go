// wsmonitor package provides monitoring Kandy application framework over web socket
//
// This package includes functions:
//	- Initialize
//	- Start
//	- Stop
/*
#########################################################################################

	Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024

	Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
						Licensed under the Apache License, Version 2.0 (the "License");
						you may not use this file except in compliance with the License.
						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

						Unless required by applicable law or agreed to in writing, software
						distributed under the License is distributed on an "AS IS" BASIS,
						WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
						See the License for the specific language governing permissions and
						limitations under the License.

	Class/module  :   WSMonitor

	Objective     :   Define the package for support web socket client connections

	This package implements the IKSHahndler Contains funtions to read and write features to connected client via
					BoradCast

#########################################################################################

	Author                 	Date        	Action      	Description

-----------------------------------------------------------------------------------------------------------------

	Ajith de Silva		24/01/2024	Created 	Created the initial version

	Ajith de Silva		29/01/2024	Updated 	Defined functions with parameters & return values

	Ajith de Silva		29/01/2024	Updated 	Implemented functions

	Ajith de Silva		01/02/2024	Added 	 	Added the status endpoint to monitor status via web socket

	Ajith de Silva		02/03/2024	Added 	 	Added the monitor endpoint to monitor activities via web socket

	Ajith de Silva		09/04/2024	Added 	 	Added the logger endpoint to monitor log entries via web socket
#########################################################################################
*/

package wsmonitor

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"text/template"
	"time"

	iappfw "agnione/v1/src/appfm/iappfw" /// import interface of ZAF
	apptypes "agnione/v1/src/appfm/types"

	wshandler "agnione.appfm/src/monitors/ws/wshandler"

	cors "agnione.appfm/src/monitors/lib"

	"github.com/gorilla/websocket"
)

// WSMonitor structure
type WSMonitor struct {
	upgrader         *websocket.Upgrader /// websocket upgrader pointer
	wsHub            *wshandler.WSHub    /// pointer to the websocket hub
	wsServer         *http.Server        /// pointer to the http server
	//wsServerExitDone *sync.WaitGroup     /// pointer to the waitgroup
	messge           chan string         /// message channel
	appInstace       iappfw.IAgniApp
	isstarted        bool /// flag to set the start/stop status

}

// Initialize initilizes the Initialize intance with default values.
// websocket.Upgrader with ReadBufferSize:1024,	WriteBufferSize:1024 and EnableCompression:true
func (wsm *WSMonitor) Initialize(appInstance iappfw.IAgniApp) {

	/// initialize the web socket upgrader
	wsm.upgrader = &websocket.Upgrader{
		ReadBufferSize:    1024,
		WriteBufferSize:   1024,
		EnableCompression: true,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},

	}
	wsm.appInstace = appInstance
	wsm.wsHub = wshandler.NewHub()           /// create an instance of websocket hub
	wsm.messge = make(chan string)           /// crete the message channel
	//wsm.wsServerExitDone = &sync.WaitGroup{} /// create the waitgroup
}

// DeInitialize clear the data and instance of objects.
func (wsm *WSMonitor) DeInitialize() {
	wsm.upgrader = nil
	wsm.messge = nil
	wsm.wsHub = nil
	wsm.wsServer = nil
	//wsm.wsServerExitDone = nil
	wsm.appInstace = nil
}

// home endpoint to get the default UI for monitor message via web socket
func (wsm *WSMonitor) home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://" + r.Host+ "/app/monitor", )
}

// status endpoint to get the default UI for the status monitor via websocket
func (wsm *WSMonitor) status(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://" + r.Host + "/app/status")
}

// status endpoint to get the default UI for the status monitor via websocket
func (wsm *WSMonitor) logger(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://" + r.Host + "/app/logger")
}

// BroadCast boradcasts a message among to the clients who connected to the monitor endpoint
func (wsm *WSMonitor) BroadCast(pMessage []byte) {
	
	if wsm != nil && wsm.wsHub != nil {
		wsm.wsHub.BroadCast(pMessage)
	}
}

// BroadCastStatus boradcsasts message to the clients who connected to the status endpoint
func (wsm *WSMonitor) BroadCastStatus(pMessage []byte) {
	//if wsm != nil && wsm.wsHub != nil {
		wsm.wsHub.BroadCastStatus(pMessage)
	//}
}

// BroadCast boradcasts a message among to the clients who connected to the monitor endpoint
func (wsm *WSMonitor) BroadCastLogEntries(pMessage []byte) {
	//if wsm != nil && wsm.wsHub != nil {
		wsm.wsHub.BroadCastLogEntries(pMessage)
	//}
}

// / MonitorClientsCount returns the number of clients who connected to monitor endpoint
func (wsm *WSMonitor) MonitorClientsCount() uint8 {
	return wsm.wsHub.MonitorClientsCount()
}

// / MonitorClientsCount returns the number of clients who connected to status endpoint
func (wsm *WSMonitor) StatusClientsCount() uint8 {
	return wsm.wsHub.StatusClientsCount()
}

// / MonitorClientsCount returns the number of clients who connected to status endpoint
func (wsm *WSMonitor) LogClientsCount() uint8 {
	return wsm.wsHub.LogClientsCount()
}

// Start starts the web socket monitoring.
// Creates a HTTP server with given address(ip/domain) and port
// Set the endpoint and handlers
// Starts the web soket hub, which manage the clients
func (wsm *WSMonitor) Start(pAddress *string, pHttp_port *int) {

	wsm.appInstace.Write2Log("starting the web socket monitor service ......", apptypes.LOG_INFO)

	go wsm.wsHub.Run() /// run the ws controller hub


	go func() {

		defer func() {
			if _r:=recover();_r!=nil{
				fmt.Println("Recovered panic ",_r)
				_r=nil
			}
			
			wsm.appInstace.Write2Log("stopped web socket monitor on " + *pAddress + ":" + strconv.Itoa(*pHttp_port) + "\r\n", apptypes.LOG_INFO)
			wsm.wsHub.Stop()
		}()

		_mux := http.NewServeMux()
		_mux.HandleFunc("/wsmonitor", wsm.home) /// set the monitor home page

		_mux.HandleFunc("/wsstatus", wsm.status) /// set the status home page

		_mux.HandleFunc("/wslogger", wsm.logger) /// set the status home page

		/// set the monitor endpoint
		_mux.HandleFunc("/app/monitor", func(w http.ResponseWriter, r *http.Request) {
			wshandler.ServeWs(wsm.wsHub, w, r)
		})

		/// set the status
		_mux.HandleFunc("/app/status", func(w http.ResponseWriter, r *http.Request) {
			wshandler.ServeWSStatus(wsm.wsHub, w, r)
		})

		/// set the logger end point
		_mux.HandleFunc("/app/logger", func(w http.ResponseWriter, r *http.Request) {
			wshandler.ServeWSLogTailer(wsm.wsHub, w, r)
		})

		wsm.appInstace.Write2Log("starting web socket monitor on " + *pAddress + ":" + strconv.Itoa(*pHttp_port) + "\r\n", apptypes.LOG_INFO)

		/// create the WS server instance with parameters
		wsm.wsServer = &http.Server{
			Addr:           *pAddress + ":" + strconv.Itoa(*pHttp_port),
			Handler:        cors.AllowAll().Handler(_mux), /// need to set origins via config
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}

		/// startes the web server on given port
		wsm.isstarted = true

		if _err := wsm.wsServer.ListenAndServe(); _err != http.ErrServerClosed {
			wsm.isstarted = false
			wsm.appInstace.Write2Log("error occurred while starting web socket monitor on " + *pAddress + ":" + 
				strconv.Itoa(*pHttp_port) + "\n" + _err.Error(), apptypes.LOG_ERROR)
			return
		}
	}()
}

/// Stop stops the web socket monitoring
func (wsm *WSMonitor) Stop() {

	if wsm.wsServer != nil {
		if _err := wsm.wsServer.Shutdown(context.TODO()); _err != nil {
			/// failure/timeout shutting down the server gracefully
			wsm.appInstace.Write2Log("web socket monitor shutdown: " +  _err.Error(), apptypes.LOG_ERROR)
		}
		wsm.wsServer = nil
	} else {
		wsm.appInstace.Write2Log("web socket monitoring serving not started", apptypes.LOG_INFO)
	}
	wsm.isstarted = false
}

func (wsm *WSMonitor) IsStarted() bool {
	return wsm.isstarted
}

var IWSMonitor WSMonitor /// set the export, IF we are to build the plun-in

/// default template for test the web socket
var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head><title>Agni App FM :: Realtime Monitoring</title><meta charset="utf-8">
<script>
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print(evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
    <h2>Agni App FM :: Websocket Monitor</h2>
    <hr>
<p>Click "Open" to create a connection to the Application and start monitoring<br>
<hr>
<form><button id="open">Open</button><button id="close">Close</button>
</form>
</td>
</tr>
<tr>
<td valign="top" width="50%">
<div id="output" style="height:500px; max-height: 70vh;overflow-y: scroll; border: 1px solid #1C6EA4;"></div>
</td></tr></table>
</body>
</html>
`))
