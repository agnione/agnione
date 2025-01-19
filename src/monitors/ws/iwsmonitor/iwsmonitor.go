// iwsmonitor interface defines the methods for websocket monitoring implementation
//
// This interface defined functions:
//	- Initialize
//	- DeInitialize
//	- Start
//	- Stop
//	- BroadCast
//	- BroadCastStatus
//	- MonitorClientsCount
//	- StatusClientsCount
/*
#########################################################################################

	Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024

	Copyright     :    © 2024 D. Ajith Nilantha de Silva contact@agnione.net
						Licensed under the Apache License, Version 2.0 (the "License");
						you may not use this file except in compliance with the License.
						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

						Unless required by applicable law or agreed to in writing, software
						distributed under the License is distributed on an "AS IS" BASIS,
						WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
						See the License for the specific language governing permissions and
						limitations under the License.

	Class/module  :   iwsmonitor

	Objective     :   Define the interface for web socket monitoring library

#########################################################################################

	Author                 	Date        	Action      	Description

-----------------------------------------------------------------------------------------------------------------

	Ajith de Silva		26/01/2004	Created 	Created the initial version

	Ajith de Silva		29/01/2004	Updated 	Defined functions with parameters & return values

	Ajith de Silva		01/02/2004	Added 		Added the status endpoint to monitor status vis web socket

	################################################################################################################
*/
package iksmonitor

// interface that needs to be implemented for Web Socket monitoring.
type IWSMonitor interface {

	// Initialize the Instance of WSMonitor
	// takes the log file path as string parameter
	Initialize()

	// DeInitialize the Instance of WSMonitor.
	DeInitialize()

	// Start starts the monitoring web socket server
	// address string parameter required for listen IP/DNS
	// port int parameter is used for listen port
	Start(address string, port int8)

	// Stop stops the web socket server which was started
	Stop()

	// BroadCast broadcasts given string message among the connected web socket clients
	BroadCast(message []byte)

	// BroadCast broadcasts given string message among the connected web socket clients
	BroadCastStatus(message []byte)

	// MonitorsCount returns the number of active web socket clients
	MonitorClientsCount() int

	// MonitorsCount returns the number of active web socket clients
	StatusClientsCount() int

	// Returns start/stop status of the Web Socket monitoring
	IsStarted() bool
}
