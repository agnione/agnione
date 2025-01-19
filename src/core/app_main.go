// agni package provides AgniOne application framework main starting point
//
// This package includes functions:
//	- Add_Routine
//	- App_Path
//	- DeInitialize
//	- Get_Context
//	- Failed_Request_Count
//	- Get_App_Info
//	- Get_App_Status
//	- Get_File_Content
//	- GetFileContentLines
//	- Get_FileInfo
//	- Get_AppUnit
//	- Get_Mailer
//	- Get_RESTClient
//	- Get_WSClient
//	- Handled_Request_Count
//	- Initialize
//	- Is_Interrupted
//	- Load_Units
//	- Memory_Usage
//	- Name
//	- Reload_Config
//	- Reload_Requested
//	- Remove_Routine
//	- Add_Request_Failed_Count
//	- Add_Request_HandleCount
//	- Routine_Count
//	- Send_Monitor_Message
//	- Start
//	- StartHttpMonitor
//	- Start_WSMonitor
//	- Started
//	- Stop
//	- Stop_WSMonitor
//	- Version
//	- WaitforClose
//	- Write2Console
//	- Write2Log
//	- WriteFileContent
/*
#########################################################################################

Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 28/10/2024

Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
						Licensed under the Apache License, Version 2.0 (the "License");
						you may not use this file except in compliance with the License.
						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

						Unless required by applicable law or agreed to in writing, software
						distributed under the License is distributed on an "AS IS" BASIS,
						WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
						See the License for the specific language governing permissions and
						limitations under the License.

Class/module  :   AgniOne Application Framework - Core Implementation

Objective     :   Define common package for Application that work as a container for the business objects.

	This package will to export plugins/libraries to the business units, so that will helps
	Business units are to implement it's features by using the framework support

#########################################################################################

	Author                 	Date        	Action      	Description
-----------------------------------------------------------------------------------------------------------------

	Ajith de Silva		26/01/2024	Created 	Created the initial version

	Ajith de Silva		28/03/2024	Updated 	added the config path and unit path as parameters

	Ajith de Silva		29/03/2024	Updated 	added the log entries broadcast via web socket

	Ajith de Silva		29/03/2024	Updated 	added the logger to application framework

#########################################################################################
*/
package agni

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"agnione/v1/src/aau/iappunit"
	iappfw "agnione/v1/src/appfm/iappfw"
	apptypes "agnione/v1/src/appfm/types"
	atypes "agnione/v1/src/appfm/types"

	"agnione.appfm/src/logger"
	ihttpm "agnione.appfm/src/monitors/http"
	iwsm "agnione.appfm/src/monitors/ws"
)

// struct to hold the framework instance data
type AgniApp struct {
	
	started      time.Time       /// holds the application started time
	wgEntries    *sync.WaitGroup /// instance wait group for go routines
	counter_lock *sync.Mutex     /// sync lock for request handle counters
	routine_lock *sync.Mutex     /// sync lock for routine counter
	info_lock *sync.Mutex     /// sync lock for info of application
	status_lock *sync.Mutex     /// sync lock for status of application
	ctx *context.Context
	WSMonitor   *iwsm.WSMonitor /// pointer web socket monitor instance
	HTTPMonitor *ihttpm.HttpMonitor
	appconfig  *apptypes.AppConfig  /// pointer for application configuration	
	coreconfig  *apptypes.FMConfig  /// pointer for application configuration

	logger     *logger.ALogger
	appUnits         []iappunit.IAppUnit /// pool to hold the application units
	
	appunit_info []atypes.AppUnitInfo
	appinfo *atypes.AppInfo
	appstatus *atypes.AppStatus
	
	last_mem_usage apptypes.MemUsage
	
	name           string /// holds the application name
	version        string /// holds the application version
	base_path      *string /// holds the base path of the application
	app_config *string /// holds the application unit base path.
	logfile_base string /// holds the base path of the log file
	app_log_file string /// holds the application log file name. Generated during the Initialize
	
	stopStatus   chan bool /// channel to control the application status readers
	stopChan   chan bool /// channel to control the application stop
	stopwsChan chan bool /// channel to control the status broadcast routine
	
	id *int 	/// holds the Process ID of the application.
	requests_handled uint64 /// holds the handled request count
	requests_failed  uint64 /// holds the failed request count
	no_of_routines   uint16   /// holds the running number of go routines
	
	reload_requested bool             /// flag to indicated application reload request
}

// Define the max pool size per appuunit
const MAX_POOL_SIZE int8 = 5

// Initialize initializes the application framework instance with default values
//
// base_path string param define the path that look for application & unit configuration path
//
// app_config_path parameter will define the path base path to application config path
// app_config string parma define the base path of the application unit files.
//

func (app *AgniApp) Initialize(pCTX_Current *context.Context, pOS_PID *int, pBase_Path *string, pApp_Config *string,pLog_Path *string,
	pREST_Port *int, pWS_Port *int) (bool, error) {
	
	app.ctx=pCTX_Current
	app.base_path = pBase_Path /// sets the base path
	app.app_config = pApp_Config
	app.id=pOS_PID

	var _err error
	_temp_path:=*pBase_Path + "config/core.config"
	app.coreconfig, _err = app.LoadCoreConfiguration(&_temp_path) /// try to load the main configuration
	if _err != nil {
		return false,errors.New("main configuration file failed to load - " +  _err.Error())
	}
	
	app.appconfig, _err = app.LoadAppConfiguration(pApp_Config) /// try to load the application configuration
	if _err != nil {
		fmt.Printf("application configuration file failed to load\n%v\n", _err)
		return false, fmt.Errorf("main configuration file failed to load - %v", _err)
	}

	app.requests_handled = 0          /// init counter request_handled
	app.requests_failed = 0           /// init counter requests_failed
	app.started = time.Now()          /// set the started time to current date-time
	
	app.stopChan = make(chan bool)    /// init the stopper channel

	
	/// init the sync locks
	app.wgEntries = &sync.WaitGroup{}
	app.counter_lock = &sync.Mutex{}
	app.routine_lock = &sync.Mutex{}
	app.status_lock=&sync.Mutex{}
	app.info_lock=&sync.Mutex{}
	
	app.appinfo=&atypes.AppInfo{}
	app.appstatus=&atypes.AppStatus{}
	
	/// set the appication name and version
	app.name = app.appconfig.App.Name
	app.version = app.appconfig.App.Version

	/// determine the log file base.
	/// first priority to input variable by App shell
	/// if not check with config file, override by the app.config then use it
	if len(*pLog_Path)>5{
		app.logfile_base=*pLog_Path
	}else{
		if len(app.coreconfig.Core.Log.File_Base_Path) > 5 {
			app.logfile_base =app.coreconfig.Core.Log.File_Base_Path
		}
	}
	
	///do the cleaning
	app.logfile_base = strings.Replace(app.logfile_base, " ", "", -1)
	if app.logfile_base == "" {
		app.logfile_base = "/var/log/app/"	/// default. if path invalid
	}

	/// set rest & ws port	
	app.coreconfig.Core.HTTPMonitor.Port=pREST_Port

	app.coreconfig.Core.WSMonitor.Port=pWS_Port
		
	///creates the logger instance and pass the parameters
	fmt.Println("Initializing the Logger with base path " +  app.logfile_base)

	app.logger = &logger.ALogger{}
	app.app_log_file = app.logfile_base + app.appconfig.App.ID + ".log"
	_, _err = app.logger.Initialize(iappfw.IAgniApp(app), app.app_log_file, apptypes.LOG_INFO, os.Getpid())
	if _err != nil {
		fmt.Println("Failed to create the instance of Logger.\n " +  _err.Error())
		app.logger = nil
	} else {
		if !app.logger.Start() {
			app.logger.DeInitialize()
			app.logger = nil
			app.Write2Console("Failed to start logger instance created with file " +  app.app_log_file)
		} else {
			app.no_of_routines++ ///increment the routine count
			app.Write2LogConsole("logger instance created with file " +  app.app_log_file, apptypes.LOG_INFO)
		}
	}

	app.Write2Log("Application " + app.name + " - " + app.version + " loaded", apptypes.LOG_INFO)

	return true, nil
}

// DeInitialize clear the initialize objects
func (app *AgniApp) DeInitialize() {

	if app.appUnits != nil {
		var _unit_count int8 = int8(len(app.appUnits))
		var _pool_index int8

		fmt.Printf("De-initializing %d AgniOne Units in the pool \n", _unit_count)
		
		
		for _pool_index = 0; _pool_index < _unit_count; _pool_index++ {
			if app.appUnits[_pool_index] != nil {
				app.appUnits[_pool_index].Deinitialize()
				fmt.Printf("App unit %d Deinitialized\n", _pool_index)
				app.appUnits[_pool_index] = nil
			}
		}
	}

	if app.WSMonitor != nil {
		app.WSMonitor.DeInitialize()
	}

	if app.HTTPMonitor != nil {
		app.HTTPMonitor.DeInitialize()
	}
	
	app.Stop_Logger() //stop the logger
	
	/// clear all varaibales - 

	app.appconfig = nil
	app.coreconfig = nil

	app.HTTPMonitor = nil
	app.WSMonitor = nil

	app.wgEntries = nil
	app.stopChan = nil
	app.stopStatus=nil
	
	app.counter_lock = nil
	app.routine_lock = nil
	app.appUnits = nil
	app.info_lock=nil
	app.status_lock=nil
	app.appunit_info=nil
	
	app.appinfo=nil
	app.appstatus=nil
	
	app.logger = nil
	app.name=""
	app.version=""
	app.base_path=nil
	app.app_config=nil
	app.logfile_base=""
	app.app_log_file=""
}

func (app *AgniApp) Get_Context() *context.Context{
	return app.ctx
}

// Reload_Config reloads the application configuration
// Returns true,nil if reload successful. Unless returns false,error
func (app *AgniApp) Reload_Config() (bool, error) {

	/* call loadAppConfiguration() and assign the result to appConfig*/
 	_temp_path:=*app.app_config + "/app.config"
	
	if _newConfig, _err := app.LoadAppConfiguration(&_temp_path); _err != nil {
		return false, _err
	} else {
		app.appconfig = _newConfig
		app.reload_requested = true /// set reload is requested. will be used in routine to reload the app
		return true, nil
	}
}


// Reload_Config reloads the application configuration
// Returns true,nil if reload successful. Unless returns false,error
func (app *AgniApp) Save_App_Config(pAppConfigData *[]byte) (bool, error) {

	_newConfig:= apptypes.AppConfig{}
	var _err error
	
	defer func(){
		_err=nil
	}()

	if _err=json.Unmarshal(*pAppConfigData,&_newConfig);_err!=nil{
		app.Write2Log("Failed to save app.config, invalid application configuration received",apptypes.LOG_ERROR)
		return false,errors.New("invalid application configuration received")
	}
	
	_temp_path:=*app.app_config + "/app.config"
	
	if _,_err=app.Write_FileContent(&_temp_path,pAppConfigData);_err!=nil{
		app.Write2Log("Failed to save " + _temp_path + ". " + _err.Error(),apptypes.LOG_ERROR)
		return false,errors.New("Failed to save " + _temp_path + ". please check error logs")
	}else{
		return true,nil
	}
}

// Reload_Config reloads the application configuration
// Returns true,nil if reload successful. Unless returns false,error
func (app *AgniApp) Reload_Requested() bool {
	return app.reload_requested
}

// App_Path returns the current application base path
func (app *AgniApp) App_Path() *string {
	return app.base_path
}

// App_Path returns the current application base path
func (app *AgniApp) Logfile_Basepath() *string {
	return &app.logfile_base
}


// App_Path returns the current application base path
func (app *AgniApp) Logfile_Name() string {
	return app.app_log_file
}

// Start starts the app framework with features defined in the configuration
func (app *AgniApp) Start() {
	
	defer func ()  {
		go func ()  {
			
				//// start update info 
				go app.update_app_info()
				go app.update_app_status()
	
				bDone:=make(chan bool,1)
				defer func ()  {
					bDone=nil
				}()
				
				go app.read_memory_usage(bDone) /// do the first memory usage read
				<- bDone
			}()
	}()
	
	app.Load_Units() /// loads the business model to run
	
	time.Sleep(time.Second * 1)
	app.stopStatus = make(chan bool)    /// init the stopper channel for status reads
		
	go app.StartHttpMonitor() /// start HTTP monitoring
	
}

func (app *AgniApp) Stop_WS_Monitor() {
	if app.WSMonitor != nil {
		if app.WSMonitor.IsStarted() {
			app.Write2LogConsole("Stopping Websocket monitoring server........", apptypes.LOG_INFO)
			app.WSMonitor.Stop()
			app.Write2LogConsole("Stopping Websocket monitoring server........ DONE", apptypes.LOG_INFO)
		}
	}
}

func (app *AgniApp) Stop_HTTPMonitor() {
	if app.HTTPMonitor != nil {
		if app.HTTPMonitor.IsStarted() {
			app.Write2LogConsole("Stopping HTTP monitoring server........", apptypes.LOG_INFO)
			app.HTTPMonitor.Stop()
			app.Write2LogConsole("Stopping HTTP monitoring server........ DONE", apptypes.LOG_INFO)
		}
	}
}

func (app *AgniApp) Stop_Units() {

	defer func(){
		recover()
		app.stopStatus=nil
	}()
	
	close(app.stopStatus)
	
	if app.appUnits != nil {
		
		app.Write2LogConsole("Found AppUnits " + strconv.Itoa(len(app.appUnits)) + " in pool", apptypes.LOG_INFO)
		
		for _index,_appUnit :=range app.appUnits {
			if _appUnit!= nil {
				if _appUnit.IsStarted() {
					app.Write2LogConsole("AppUnit - [" + strconv.Itoa(_index) + "] Stop called" , apptypes.LOG_INFO)
					_appUnit.Stop()
					time.Sleep(time.Millisecond * 200)
				}else{
					app.Write2LogConsole("AppUnit - [" + strconv.Itoa(_index) + "] NOT Started", apptypes.LOG_INFO)
				}
				_appUnit=nil
				app.appUnits[_index]=nil
			}
		}
		
		clear(app.appunit_info)
	}else{
		app.Write2LogConsole("No AppUnits loaded", apptypes.LOG_INFO)
	}
}

func (app *AgniApp) Stop_Logger() {

	if app.logger != nil {
		if app.logger.IS_Started {
			fmt.Println("Stopping the logger........")
			app.logger.Stop()
			fmt.Println("Stopping the logger ........ DONE")
			app.no_of_routines-- ///reduce the routine count
		}
	}

}

// Stop flags the framework to stop by setting the flag to channel
// signal is sent by the app shell.
func (app *AgniApp) Stop() {

	app.Write2LogConsole("Application Framework stopped called......", apptypes.LOG_INFO)

	app.Stop_Units()

	app.Stop_WS_Monitor()

	app.Stop_HTTPMonitor()

	app.Write2LogConsole("Application Framework stopped called........ DONE", apptypes.LOG_INFO)

}

// Terminate closes the main stopper channel of the framework
func (app *AgniApp) Terminate() {

	close(app.stopChan) //stop main channel

}

// / loads the Application Units Model that implements the logic/code base on the business requirements
func (app *AgniApp) Load_Units() {

	app.appUnits = make([]iappunit.IAppUnit, 0) /// creates the pool of AppUnits
	
	var _unitIndex int=0
	var _appUnit apptypes.Appunit

	for _unitIndex, _appUnit = range app.appconfig.Appunits {

		///if the app unit is disabled, log it and continue the the next
		if _appUnit.Enable == 0 {
			app.Write2LogConsole(strconv.Itoa(_unitIndex) + " unit " + _appUnit.Uname + " -> DISABLED", apptypes.LOG_WARN)
		} else {
			app.Load_AppUnit(&_unitIndex, &_appUnit)
		}
	}

	if len(app.appUnits) == 0 {
		app.appUnits = nil
		app.Write2LogConsole("No AgniOne Units loaded", apptypes.LOG_WARN)

	} else {
		app.Write2LogConsole("Loaded & started " +  strconv.Itoa(len(app.appUnits))  + " pool of appunits", apptypes.LOG_INFO)
		
		app.appunit_info=make([]atypes.AppUnitInfo, len(app.appUnits))
	}
}

// / loads the Applcation Unit(s) Model that implemets the logic/code base on the business requirments
func (app *AgniApp) Load_AppUnit(_unitIndex *int, appunit *apptypes.Appunit) *int8 {
	
	var _loaded_count int8 = 0
	
	defer func ()  {
		if _r:=recover();_r!=nil{
			fmt.Printf("Recovered from Load_Unit panic %v",_r)
			_r=nil
		}
	}()
	
	if appunit.PoolSize == 0 {
		app.Write2LogConsole("Failed to load AgniOne Unit " +  appunit.Uname + " - pool size set to 0", apptypes.LOG_WARN)
		return &_loaded_count
	}

	//do check for defined pool size and control it
	if appunit.PoolSize > MAX_POOL_SIZE {
		appunit.PoolSize = MAX_POOL_SIZE
	}

	app.Write2LogConsole("Found " + strconv.Itoa(int(appunit.PoolSize)) + " pool setting for appunit " + appunit.Uname, apptypes.LOG_INFO)
	
	var _pool_index int8
	
	/// loads the app unit into pool
	for _pool_index = 0; _pool_index < appunit.PoolSize; _pool_index++ {
		_appUnit, _err := app.Get_AppUnit(_unitIndex) /// Get the AgniOne Unit instance

		if _err != nil {
			
			app.Write2LogConsole("Failed to load AgniOne " + appunit.Uname  + " - " + appunit.Path + ". " + _err.Error(), apptypes.LOG_ERROR)
			return &_loaded_count
		}
		
		app.Write2LogConsole("Loading the " + strconv.Itoa(int(_pool_index)) + " appunit of pool [" + strconv.Itoa(int(appunit.PoolSize)) +"] of " + appunit.Uname, apptypes.LOG_INFO)
		
		if _, _err = _appUnit.Initialize(app, len(app.appUnits), appunit.Uname,  appunit.Path,appunit.ConfigFile); _err != nil {
			
			app.Write2LogConsole(fmt.Sprintf("Failed to initialize " + strconv.Itoa(int(_pool_index) +1) + "instace of the appunit " + strconv.Itoa(int(appunit.PoolSize)) + 
					" into pool of " + appunit.Uname + "\n" + _err.Error(),  appunit.PoolSize, appunit.Uname, _err), apptypes.LOG_ERROR)

			_appUnit = nil
			return &_loaded_count
		}
		
		app.Write2LogConsole("Starting (" + strconv.Itoa(int(_pool_index)) + ") of [" + strconv.Itoa(int(appunit.PoolSize)) +"] - " + 
			appunit.Uname + ".............\nAppunit info -> " + _appUnit.Info().BuildGoVersion, apptypes.LOG_INFO)
		
		if _, _err = _appUnit.Start(); _err != nil {

			/// failed to start. clear the resources
			app.Write2LogConsole("Failed to start (" + strconv.Itoa(int(_pool_index)) + ") of [" + strconv.Itoa(int(appunit.PoolSize)) +"] - " + appunit.Uname + ". " + _err.Error(), apptypes.LOG_ERROR)

			_appUnit.Deinitialize()
			_appUnit = nil
			continue
		} else {
			/// All good. store the started AppUnit in the pool
			_loaded_count++
			app.appUnits = append(app.appUnits, _appUnit)
			
			app.Write2LogConsole("Started ------ (" + strconv.Itoa(int(_pool_index)) + ") of [" + strconv.Itoa(int(appunit.PoolSize)) +"] - " + appunit.Uname + " successfully", apptypes.LOG_INFO)
			
		}
	}

	app.Write2LogConsole(
		"Started\t" + strconv.Itoa(int(_loaded_count)) + " out of " + strconv.Itoa(len(app.appUnits)) + " pool of appunit " + appunit.Uname + "\n",apptypes.LOG_INFO)

	return &_loaded_count
}

// Is_Interrupted returns if the applcation stop channel.
// External package routines should check this for an interupption/stop request and stop accordingly.
func (app *AgniApp) Is_Interrupted() chan bool {
	return app.stopChan
}

// WaitforClose waits for all the routines close
func (app *AgniApp) WaitforClose() {
	app.wgEntries.Wait()
}

// Add_Routine increment the routine count and add 1 to the framework waitgroup
func (app *AgniApp) Add_Routine() {
	app.routine_lock.Lock()
	defer app.routine_lock.Unlock()
	app.no_of_routines++
	app.wgEntries.Add(1)
}

// Remove_Routine decrements the routine count and remove 1 from the framework waitgroup
func (app *AgniApp) Remove_Routine() {
	app.routine_lock.Lock()
	defer app.routine_lock.Unlock()
	app.no_of_routines--
	app.wgEntries.Done()
}

var IAgniApp AgniApp
