//
//#################################################################################################################
// Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024
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
// Class/module  :   Kandy Application Framework - Core Monitor Implementation
// Objective     :   Define common package for Application that work as a container for the business objects.
//					This package will to export plugins/libraries to the business objects, so that will helps
//					business objects to implements it's features by using the framework support
//################################################################################################################
// Author                        Date        Action      Description
//------------------------------------------------------------------------------------------------------
// Ajith de Silva				26/01/2024	Created 	Created the initial version
// Ajith de Silva				29/01/2024	Updated 	Defined functions with parameters & return values
// Ajith de Silva				29/01/2024	Updated 	Updated the SSEMonitor as library
//#################################################################################################################
//

package agni

import (
	apptypes "agnione/v2/src/appfm/types"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	ihttpm "agnione.appfm/src/monitors/http"
	issem "agnione.appfm/src/monitors/sse/ssehandler"
)

/* ###################### START ###### 		Monitor related functions ###################### */

// StartHttpMonitor starts the HTTP monitoring.
// HTTP server expose multiple endpoints to view application details and also manage application
// up to some extend.
//
// /info - returns the application status information format of apptypes.AppInfo
//
// /status - returns the application status information format of apptypes.AppStatus
//
// /live - returns the application status LIVE of not
//
// /admin/log - returns the last 100 rows of the application log
//
// /admin/monitor/start - starts the web socket monitoring. (if not already started)
//
// /admin/monitor/stop - stops the web socket monitoring. (if already started)
//
// /admin/config/reload - reloads application configuration
func (app *AgniApp) StartHttpMonitor() {

	defer func() {
		recover()
	}()

	app.Write2LogConsole("starting HTTP monitoring ......", apptypes.LOG_INFO)

	app.HTTPMonitor = &ihttpm.HttpMonitor{}
	app.HTTPMonitor.Initialize(app)
	app.HTTPMonitor.Start(&app.coreconfig.Core.Monitor.Host, app.coreconfig.Core.Monitor.Port)

	app.Write2LogConsole("HTTP monitoring started on "+app.coreconfig.Core.Monitor.Host+":"+
		strconv.Itoa(*app.coreconfig.Core.Monitor.Port), apptypes.LOG_INFO)

	app.SSEMonitor = app.HTTPMonitor.Get_SSEMonitorPtr()

	app.Add_Routine()
	go app.broadcast_log_messages()

	app.Add_Routine()
	go app.broadcast_status_messages()

	app.Add_Routine()
	go app.broadcast_event_messages()

}

// Send_Event_Message broadcasts given message via SSE monitoring.
func (app *AgniApp) Send_Event(pMessage string) {
	app.event_message <- pMessage
}

// broadcast_status broadcast the application status via sse monitoring
func (app *AgniApp) broadcast_status_messages() {

	time.Sleep(2 * time.Second) /// set delay to init

	defer func() {
		if _r := recover(); _r != nil {
			fmt.Println("Recovered panic from broadcast_status. ", _r)
			_r = nil
		}

		app.Remove_Routine()
		app.Write2Log("broadcasting status via SSE stopped", apptypes.LOG_INFO)
	}()

	if app.SSEMonitor == nil {
		return
	}

	var _statusmsg []byte
	_ticker := time.NewTicker(5 * time.Second)

	defer func() {
		_ticker.Stop()
		_ticker = nil
		_statusmsg = nil
		app.Write2Log("stopping the status broadcasting SSE", apptypes.LOG_INFO)
	}()

	app.Write2Log("broadcasting status via SSE started", apptypes.LOG_INFO)
	_sseEvent := issem.SSE_Event{}
	for {
		select {
		case <-app.stopChan:
			return
		case <-_ticker.C:
			{

				if app.SSEMonitor == nil {
					return
				}

				if app.SSEMonitor.StatusClientsCount() > 0 {
					if _statusmsg, _ = json.Marshal(app.Get_App_Status()); _statusmsg != nil {
						_sseEvent.ID = time.Microsecond.String()
						_sseEvent.Message = string(_statusmsg)
						app.SSEMonitor.Broadcast_Status(_sseEvent) /// broadcast the received status
					}
				}
			}
		}
	}
}

func (app *AgniApp) broadcast_event_messages() {

	time.Sleep(2 * time.Second) /// set delay to init

	defer func() {
		if _r := recover(); _r != nil {
			fmt.Println("Recovered panic from broadcast event messages. ", _r)
			_r = nil
		}

		app.Remove_Routine()
		app.Write2Log("broadcast event messages via SSE stopped", apptypes.LOG_INFO)
	}()

	if app.SSEMonitor == nil {
		return
	}

	app.Write2Log("broadcasting event messages via SSE started", apptypes.LOG_INFO)
	_sseEvent := issem.SSE_Event{}
	for {
		select {
		case <-app.stopChan:
			return
		case _sseEvent.Message = <-app.event_message:
			{
				if app.SSEMonitor == nil {
					return
				}
				if app.SSEMonitor.MonitorClientsCount() > 0 && len(_sseEvent.Message) > 0 {
					_sseEvent.ID = time.Microsecond.String()
					app.SSEMonitor.Broadcast_Event(_sseEvent) /// broadcast the received status
				}
			}
		}
	}
}

func (app *AgniApp) broadcast_log_messages() {

	time.Sleep(2 * time.Second) /// set delay to init

	defer func() {
		if _r := recover(); _r != nil {
			fmt.Println("Recovered panic from broadcast log messages. ", _r)
			_r = nil
		}

		app.Remove_Routine()
		app.Write2Log("broadcast log messages via SSE stopped", apptypes.LOG_INFO)
	}()

	if app.SSEMonitor == nil {
		return
	}

	app.Write2Log("broadcasting log messages via SSE started", apptypes.LOG_INFO)
	_sseEvent := issem.SSE_Event{}
	for {
		select {
		case <-app.stopChan:
			return
		case _sseEvent.Message = <-app.log_message:
			{
				if app.SSEMonitor == nil {
					return
				}

				if app.SSEMonitor.MonitorClientsCount() > 0 && len(_sseEvent.Message) > 0 {
					_sseEvent.ID = time.Microsecond.String()
					app.SSEMonitor.Broadcast_Log(_sseEvent) /// broadcasts received log message
				}
			}
		}
	}
}
