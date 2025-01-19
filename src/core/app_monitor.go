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
// Ajith de Silva				29/01/2024	Updated 	Updated the WSMonitor as library
//#################################################################################################################
//

package agni

import (
	apptypes "agnione/v1/src/appfm/types"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	ihttpm "agnione.appfm/src/monitors/http"
	iwsm "agnione.appfm/src/monitors/ws"
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
	
	defer func ()  {
		recover()
	}()

	if app.coreconfig.Core.HTTPMonitor.Enable == 0 {
		app.Write2LogConsole("HTTP monitoring is disabled", apptypes.LOG_INFO)
		return
	}

	
	app.Write2LogConsole("starting HTTP monitoring ......", apptypes.LOG_INFO)
	
	app.HTTPMonitor = &ihttpm.HttpMonitor{}
	app.HTTPMonitor.Initialize(app)
	app.HTTPMonitor.Start(&app.coreconfig.Core.HTTPMonitor.Host, app.coreconfig.Core.HTTPMonitor.Port)
	
	app.Write2LogConsole("HTTP monitoring started on " + app.coreconfig.Core.HTTPMonitor.Host+ ":" + 
		strconv.Itoa(*app.coreconfig.Core.HTTPMonitor.Port), apptypes.LOG_INFO)
}

// Start_WSMonitor starts the web socket monitor by using the given configuration in the main config.
// Returns true,nil if started without issue. Unless returns false,error

func (app *AgniApp) Start_WSMonitor() (bool, error) {

	defer func ()  {
		recover()
	}()
	
	if app.WSMonitor != nil {
		return false, errors.New("websocket monitoring already started")
	}
	/// check if ws monitoring is enabled
	if app.coreconfig.Core.WSMonitor.Enable == 0 {
		app.Write2LogConsole("web socket monitoring is disabled", apptypes.LOG_WARN)
		return false,  errors.New("failed to start web socket monitoring. web socket monitoring is disabled")
	}
	
	app.WSMonitor = &iwsm.WSMonitor{} /// creates the instance of WSMonitor
	app.WSMonitor.Initialize(app)     /// initialize it
	/// starts the web socket monitor with ip and port given in the configuration file
	app.WSMonitor.Start(&app.coreconfig.Core.WSMonitor.Host, app.coreconfig.Core.WSMonitor.Port)
	time.Sleep(3 * time.Second)

	if app.WSMonitor.IsStarted() {
		app.stopwsChan = make(chan bool) /// creates a channel to control the broadcast routine
		app.Write2LogConsole("websocket monitoring started on " + app.coreconfig.Core.WSMonitor.Host + ":" + strconv.Itoa(*app.coreconfig.Core.WSMonitor.Port), apptypes.LOG_INFO)

		app.Add_Routine()          /// increase the routine count
		go app.broadcast_status() /// starts the status broadcast routine

		return true, nil
	} else {
		app.Write2LogConsole("failed to start websocket monitoring started on "+ app.coreconfig.Core.WSMonitor.Host + ":" + strconv.Itoa(*app.coreconfig.Core.WSMonitor.Port), apptypes.LOG_ERROR)

		app.WSMonitor.DeInitialize()
		app.WSMonitor = nil
		return false, errors.New("failed to start web socket monitoring")
	}
}

// Stop_WSMonitor stops the web socket monitoring
// Returns true if there is no error
func (app *AgniApp) Stop_WSMonitor() bool {

	if app.WSMonitor != nil {
		close(app.stopwsChan) /// close the ws broadcast channel
		time.Sleep(2 * time.Second) /// wait broadcast routines to stop
		app.WSMonitor.Stop()
		app.WSMonitor.DeInitialize()
		app.WSMonitor = nil
	}
	return true
}

// broadcast_status broadcast the application status via web socket monitoring
func (app *AgniApp) broadcast_status() {

	time.Sleep(2 * time.Second) /// set delay to init & set the websocket server

	defer func() {
		if _r:=recover();_r!=nil{
			fmt.Println("Recovered panic ",_r)
			_r=nil
		}
		
		app.Remove_Routine()
		app.Write2Log("broadcasting status via web socket stopped", apptypes.LOG_INFO)
	}()

	if app.WSMonitor == nil {
		return
	}

	var _statusmsg []byte
	_ticker := time.NewTicker(1 * time.Second)
	
	defer func ()  {
		_ticker.Stop()
		_ticker=nil
		_statusmsg=nil
		app.Write2Log("stopping the status broadcasting via web socket", apptypes.LOG_INFO)
	}()

	app.Write2Log("broadcasting status via web socket started", apptypes.LOG_INFO)
	for {
		select {
		case <-app.stopwsChan:
			return
		case <-app.stopChan:
			return
		case  <-_ticker.C:{
			if app.WSMonitor.StatusClientsCount() > 0 {
				if _statusmsg, _ = json.Marshal(app.Get_App_Status()); _statusmsg != nil {
					app.WSMonitor.BroadCastStatus(_statusmsg) /// broadcast the received status
					_statusmsg = nil
				}
			}
		}
		}
	}
}

// Send_Monitor_Message broadcasts given message via wbe socket monitoring.
// If the web socket monitoring is not started then this message will be discarded.
func (app *AgniApp) Send_Monitor_Message(pMessage []byte) {

	go func (pMessage []byte)  {
		defer recover()
		
		if len(pMessage) == 0 || app.WSMonitor == nil {
			return
		}

		if app.WSMonitor.MonitorClientsCount() > 0 {
			app.WSMonitor.BroadCast((pMessage))
		}
	}(pMessage)
}
