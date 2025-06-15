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
// Class/module  :   AgniOne Application Framework - Core Logger Implementation
//
// Objective     :   Define common package for Application that work as a container for the business objects.
//					This package will to export plugins/libraries to the business objects, so that will helps
//					business objects to implements it's features by using the framework support
//################################################################################################################
// Author                        Date        Action      Description
//------------------------------------------------------------------------------------------------------
// Ajith de Silva				26/01/2024	Created 	Created the initial version
// Ajith de Silva				29/01/2024	Updated 	Defined functions with parameters & return values
// Ajith de Silva				29/01/2024	Added 		Added the Write2Log method with log level parameter
// Ajith de Silva 				09/04/2024  Optimized   optimized the write log function
// 														Added the log message broadcast to log function
//#################################################################################################################

package agni

import (
	aftypes "agnione/v2/src/appfm/types"
	"fmt"

	"agnione.appfm/src/logger"
)

// Write2Console writes the given entry into console
func (app *AgniApp) Write2Console(pEntry string) {
	fmt.Println(pEntry)
}

func (app *AgniApp) Set_LogLevel(log_level aftypes.LogLevel) {
	app.logger.Set_LogLevel(log_level)
}

// Write2Log writes the given entry into the application log
func (app *AgniApp) Write2Log(pEntry string, pLog_Level aftypes.LogLevel) {
	app.log_entry <- logger.LogMessage{Msg_Entry: pEntry, Msg_Type: pLog_Level}
}

// Write2Log writes the given entry into the application log
func (app *AgniApp) Write2LogConsole(pEntry string, pLog_Level aftypes.LogLevel) {
	fmt.Println(pEntry)
	app.log_entry <- logger.LogMessage{Msg_Entry: pEntry, Msg_Type: pLog_Level}
}

func (app *AgniApp) log_writer() {

	defer func() {
		if _r := recover(); _r != nil {
			fmt.Println("Recovered panic from log_writer. ", _r)
		}

		app.Remove_Routine()
		app.Write2Console("log_writer stopped")
	}()

	_logEntry := logger.LogMessage{}

	defer func() {
		_logEntry = logger.LogMessage{}
	}()

	app.Write2Console("log_writer started")

	for {
		select {
		case <-app.stopChan:
			return
		case _logEntry = <-app.log_entry:
			{
				if len(_logEntry.Msg_Entry) > 0 {

					if app.logger == nil {
						fmt.Println(_logEntry.Msg_Entry)
						continue
					} else {
						app.logger.WriteLog(_logEntry)
					}

					if app.SSEMonitor != nil {
						if app.SSEMonitor.LogClientsCount() > 0 {
							app.log_message <- _logEntry.Msg_Entry /// broadcasts received log message
						}
					}
				}
			}
		}
	}
}
