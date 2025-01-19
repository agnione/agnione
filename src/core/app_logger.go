//################################################################################################################
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
//
// Class/module  :   Kandy Application Framework - Core Logger Implementation
//
// Objective     :   Define common package for Application that work as a container for the business objects.
//					This package will to export plugins/libraries to the business objects, so that will helps
//					business objects to implements it's features by using the framework support
//################################################################################################################
// Author                        Date        Action      Description
//------------------------------------------------------------------------------------------------------
// Ajith de Silva				26/01/2024	Created 	Created the initial version
// Ajith de Silva				29/01/2024	Updated 	Defined functions with parameters & return values
// Ajith de Silva				29/01/2024	Added 		Added the Write2Log method with loglevel parameter
// Ajith de Silva 				09/04/2024  Optimized   optimized the write log function
// 														Added the log message broadcast to log function
//#################################################################################################################

package agni

import (
	aftypes "agnione/v1/src/appfm/types"
	"fmt"
	"time"

	"agnione.appfm/src/logger"
)

// Write2Console writes the given entry into console
func (app *AgniApp) Write2Console(pEntry string) {
	go func ()  {
		fmt.Println(pEntry)
	}()
}

func (app *AgniApp) Set_LogLevel(log_level aftypes.LogLevel) {
	app.logger.Set_LogLevel(log_level)
}


// Write2Log writes the given entry into the application log
func (app *AgniApp) Write2Log(pEntry string, pLog_Level aftypes.LogLevel) {

	go func ()  {
		/// broadcast log entries to websocket endpoint	
		if app.WSMonitor != nil && app.WSMonitor.IsStarted() {
			go app.WSMonitor.BroadCastLogEntries([]byte(time.Now().Format("2006-01-02 - 15:04:05") + pEntry))
		}	
		

		/// in case we do not have logger initialized. we write the entries to the console
		if app.logger == nil {
			fmt.Println(pEntry)
			return
		}
		///pass the log entry to logger
		app.logger.WriteLog(logger.LogMessage{Msg_Entry: pEntry, Msg_Type: pLog_Level})
		}()
}

// Write2Log writes the given entry into the application log
func (app *AgniApp) Write2LogConsole(pEntry string, pLog_Level aftypes.LogLevel) {
	go func(){
		app.Write2Console(pEntry)
		app.Write2Log(pEntry,pLog_Level)
	}()
}