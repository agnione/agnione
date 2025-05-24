//
//##################################################################################################################
// Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024
//
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
// Class/module  :   Kandy Application Framework - Core Status Implementation
//
// Objective     :   Define common package for Application that work as a container for the business objects.
//					This package will to export plugins/libraries to the business objects, so that will helps
//					business objects to implements it's features by using the framework support
//################################################################################################################
// Author                        Date        Action      Description
//------------------------------------------------------------------------------------------------------
// Ajith de Silva				26/01/2024	Created 	Created the initial version
// Ajith de Silva				29/01/2024	Updated 	Defined functions with parameters & return values
// Ajith de Silva				29/01/2024	Updated 	Updated the WSMonitor as library
// Ajith de Silva				08/03/2024	Updated 	Optimized routine sync lock
//#################################################################################################################
///

package agni

import (
	"agnione/v2/src/aau/iappunit"
	atypes "agnione/v2/src/appfm/types"
	"fmt"
	"runtime"
	"sync/atomic"
	"time"

	kutls "agnione.appfm/src/utils"
)

// Name returns the application name
func (app *AgniApp) Name() string {
	return app.name
}

// ID returns the application name
func (app *AgniApp) ID() string {
	return app.appconfig.App.ID
}

// PID returns the application process ssID
func (app *AgniApp) PID() int {
	return *app.id
}

// Version returns the application version
func (app *AgniApp) Version() string {
	return app.version
}

// Memory_Usage returns the current application current memory usage
func (app *AgniApp) Memory_Usage() string {
	_mem := &runtime.MemStats{}

	defer func() {
		_mem = nil
	}()

	runtime.ReadMemStats(_mem)
	_mem_usage := kutls.FormatByteSize(_mem.Alloc)

	return _mem_usage
}

// Routine_Count returns the number of routines currently running
func (app *AgniApp) Routine_Count() uint32 {
	return atomic.LoadUint32(&app.no_of_routines)
}

// Failed_Request_Count returns the number of failed requests count set by Add_Request_Failed_Count().
func (app *AgniApp) Failed_Request_Count() uint64 {
	return atomic.LoadUint64(&app.requests_failed)
}

// Add_Request_HandleCount adds 1 to the request handle count.
// This function is useful to external modules to update his request handle count
func (app *AgniApp) Add_Request_HandleCount() {
	atomic.AndUint64(&app.requests_handled, 1)
}

// Add_Request_Failed_Count adds 1 to the request failed handle count.
// This function is useful to external modules to update his request handle count
func (app *AgniApp) Add_Request_Failed_Count() {
	atomic.AndUint64(&app.requests_failed, 1)
}

// Handled_Request_Count returns the number of requests handled set by Add_Request_HandleCount().
func (app *AgniApp) Handled_Request_Count() uint64 {
	return atomic.LoadUint64(&app.requests_handled)
}

// Started returns the application started time
func (app *AgniApp) Started() time.Time {
	return app.started
}

// Get_App_Status returns the current application status as [aypes.AppStatus]
func (app *AgniApp) Get_App_Status() atypes.AppStatus {
	return *app.appstatusPtr.Load()
}

func (app *AgniApp) Get_App_Info() atypes.AppInfo {
	return *app.appInfoPtr.Load()
}

func (app *AgniApp) Get_Ready_Status() *atypes.AppReady_Status {

	app.ready_status_lock.RLock()
	defer app.ready_status_lock.RUnlock()

	for k := range app.appready_status.Status {
		delete(app.appready_status.Status, k)
	}

	app.appready_status.Status = make(map[string]bool, len(app.appUnits))

	for _, _appUnit := range app.appUnits {
		app.appready_status.Status[_appUnit.Status().Info.Name] = _appUnit.IsReady()
	}

	return app.appready_status
}

func (app *AgniApp) update_units_info(pDoneChan chan bool) {
	defer recover()

	var _unit iappunit.IAppUnit
	var _index int

	defer func() {
		_unit = nil
	}()

	for _index, _unit = range app.appUnits {
		app.appunit_info[_index] = *_unit.Status()
	}

	pDoneChan <- true
}

func (app *AgniApp) read_memory_usage(pDoneChan chan bool) {

	var _currentMem runtime.MemStats
	runtime.ReadMemStats(&_currentMem)

	if _currentMem.Alloc > app.last_mem_usage.Heap {
		app.last_mem_usage.Heap = _currentMem.Alloc - app.last_mem_usage.Heap
	} else {
		app.last_mem_usage.Heap = app.last_mem_usage.Heap - _currentMem.Alloc
	}
	if _currentMem.HeapAlloc > app.last_mem_usage.HeapAlloc {
		app.last_mem_usage.HeapAlloc = _currentMem.HeapAlloc - app.last_mem_usage.HeapAlloc
	} else {
		app.last_mem_usage.HeapAlloc = app.last_mem_usage.HeapAlloc - _currentMem.HeapAlloc
	}

	if _currentMem.TotalAlloc > app.last_mem_usage.Total {
		app.last_mem_usage.Total = _currentMem.TotalAlloc - app.last_mem_usage.Total
	} else {
		app.last_mem_usage.Total = app.last_mem_usage.Total - _currentMem.TotalAlloc
	}

	pDoneChan <- true
}

func (app *AgniApp) update_app_status() {

	bDone := make(chan bool, 1)

	defer func() {
		bDone = nil
	}()

	go app.read_memory_usage(bDone) /// call memory usage read

	app.appstatus.Req_Handled = app.Handled_Request_Count()
	app.appstatus.Req_Failed = app.Failed_Request_Count()

	app.appstatus.Routines = app.Routine_Count()

	<-bDone
	app.appstatus.Mem_Usage = app.last_mem_usage

	if app.SSEMonitor != nil {
		app.appstatus.MonitorClients = app.SSEMonitor.MonitorClientsCount()
		app.appstatus.StatusClients = app.SSEMonitor.StatusClientsCount()
	}

	app.appstatusPtr.Store(app.appstatus)
}

// Get_App_Info returns the current application information as [ztypes.AppInfo]
func (app *AgniApp) update_app_info() {

	bDone := make(chan bool, 1)

	defer func() {
		bDone = nil
	}()

	go app.update_units_info(bDone) /// call unit info read

	app.appinfo.Name = app.name
	app.appinfo.Version = app.version
	app.appinfo.Started = app.Started().Format(time.RFC822)
	app.appinfo.PID = *app.id

	if app.HTTPMonitor != nil {
		app.appinfo.HTTPMonitor_Started = app.HTTPMonitor.IsStarted()
	}

	if app.SSEMonitor != nil {
		app.appinfo.SSEMonitor_Started = app.SSEMonitor.IsStarted()
	}

	<-bDone
	app.appinfo.AppUnits = app.appunit_info

	app.appinfo.UpTime = fmt.Sprintf("%.f", time.Since(app.Started()).Seconds())

	app.appInfoPtr.Store(app.appinfo)
}

func (app *AgniApp) Update_Status_Process() {

	defer func() {
		recover()
		app.Remove_Routine()
	}()

	_ticker := time.NewTicker(time.Second * 5)

	defer func() {
		_ticker.Stop()
		_ticker = nil
	}()

	for {
		select {
		case <-app.stopChan:
			return
		case <-app.stopStatus:
			return
		case <-_ticker.C:
			app.update_app_status()
		}
	}
}

func (app *AgniApp) Update_Info_Process() {
	defer func() {
		recover()
		app.Remove_Routine()
	}()

	_ticker := time.NewTicker(time.Second * 5)

	defer func() {
		_ticker.Stop()
		_ticker = nil
	}()

	for {
		select {
		case <-app.stopChan:
			return
		case <-app.stopStatus:
			return
		case <-_ticker.C:
			app.update_app_info()
		}
	}
}
