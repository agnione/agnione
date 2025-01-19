/*
#########################################################################################

	Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024	Copyright     :   contact@agnione.net
	Class/module  :   httpmonitor
	Objective     :   Ability to monitor and control the application via HTTP/REST protocol

#########################################################################################

	Author                 	Date        	Action      	Description

------------------------------------------------------------------------------------------------------

	Ajith de Silva		29/01/2024	Created 	Created the initial version

	Ajith de Silva		29/01/2024	Updated 	Defined functions with parameters & return values

	Ajith de Silva		29/01/2024	Added 		Added multiple API endpoints

#########################################################################################
*/
package httmonitor

import (
	iappfw "agnione/v1/src/appfm/iappfw" /// import interface of ZAF
	apptypes "agnione/v1/src/appfm/types"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	cors "agnione.appfm/src/monitors/lib"
)

// RESPONSE_FORMAT define the default response format to application/JSON
const RESPONSE_FORMAT = "application/json"


// HttpMonitor struct of the HttpMonitor
type HttpMonitor struct {
	appInstance         iappfw.IAgniApp
	httpServerExitDone *sync.WaitGroup
	apikeys            *[]string
	httpServer         *http.Server
	isstarted          bool

}

// Initialize initializes the HttpMonitor instance
// Requires the IZApp interface as parameter
func (hm *HttpMonitor) Initialize(pApp_Instance iappfw.IAgniApp) {
	hm.appInstance = pApp_Instance
	hm.httpServerExitDone = &sync.WaitGroup{}
	
	var _err error
	
	var _temp_path=*hm.appInstance.App_Path() + "config/apikeys.config"
	/// load the api keys to REST authentication
	hm.apikeys, _err = hm.appInstance.Get_FileContent_Lines(&_temp_path)
	if _err != nil {
		hm.appInstance.Write2Log("HTTP Monitor :: failed to read the " + _temp_path + " - " +  _err.Error(), apptypes.LOG_ERROR)
	}
	_temp_path=""
}

func (hm *HttpMonitor) DeInitialize() {
	hm.appInstance = nil
	hm.apikeys = nil

	hm.httpServerExitDone = nil
	hm.httpServer = nil
}


// authMiddleware checks if the apikey is given in the HTTP request header
//
// Call next handler function if given API key is valid
// If not valid then send Unauthorized response to the client
func (hm *HttpMonitor) authMiddleware(pNext http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !hm.is_authorized(r) {
			hm.setJsonResp([]byte(""), http.StatusUnauthorized, w)
			return
		} else {
			pNext.ServeHTTP(w, r)
		}
	})
}

// is_authorized checks if the apikey is given in the HTTP request header
//
// if given apikey is valid then returns true
// Unless returns false
func (hm *HttpMonitor) is_authorized(pRequest *http.Request) bool {

	_apiKey := pRequest.Header.Get("apikey")
	if _apiKey == "" {
		return false
	}

	var _value string
	for _, _value = range *hm.apikeys {
		if _value == _apiKey {
			return true
		}
	}
	
	return false
}

// Start starts the HttpMonitor on given address and port
// Parameter address is the host address
// Parameter http_port is the port to bind on given address
func (hm *HttpMonitor) Start(pAddress *string, pHttp_Port *int) {

	hm.httpServerExitDone.Add(1)

	go func() {

		defer func() {
			if _r:=recover();_r!=nil{
				fmt.Println("Recovered panic ",_r)
				_r=nil
			}
			
			hm.httpServer = nil
			hm.httpServerExitDone.Done()
			hm.isstarted = false
		}()

		_mux := http.NewServeMux()

		_mux.HandleFunc("/live", hm.live)

		///apikey authentication have been set to below routes
		_mux.Handle("/info", hm.authMiddleware(http.HandlerFunc(hm.info)))
		_mux.Handle("/status", hm.authMiddleware(http.HandlerFunc(hm.status)))
		_mux.Handle("/admin/monitor/start", hm.authMiddleware(http.HandlerFunc(hm.startmonitor)))
		_mux.Handle("/admin/monitor/stop", hm.authMiddleware(http.HandlerFunc(hm.stopmonitor)))
		_mux.Handle("/admin/config/reload", hm.authMiddleware(http.HandlerFunc(hm.config_reload)))
		_mux.Handle("/admin/config/save", hm.authMiddleware(http.HandlerFunc(hm.config_save)))

		/// sets the log level at runtime
		_mux.Handle("/admin/log/setlevel", hm.authMiddleware(http.HandlerFunc(hm.set_log_level)))
		
		/// application units management
		_mux.Handle("/admin/units", hm.authMiddleware(http.HandlerFunc(hm.list_units)))
		_mux.Handle("/admin/unit/stop", hm.authMiddleware(http.HandlerFunc(hm.stop_unit)))
		_mux.Handle("/admin/unit/{name}/start", hm.authMiddleware(http.HandlerFunc(hm.start_unit)))
		_mux.Handle("/admin/unit/{name}/restart?force=", hm.authMiddleware(http.HandlerFunc(hm.restart_unit)))
		_mux.Handle("/admin/unit/{name}/status", hm.authMiddleware(http.HandlerFunc(hm.status_unit)))
	
		hm.isstarted = true
	
		/// need to set origins via config
		hm.httpServer = &http.Server{Addr: fmt.Sprintf("%s:%d", *pAddress, *pHttp_Port), Handler: cors.AllowAll().Handler(_mux)}
		if _err := hm.httpServer.ListenAndServe(); _err != http.ErrServerClosed {
			hm.appInstance.Write2Log(fmt.Sprintf("HTTP Monitor API failed to start on port %d. %v", *pHttp_Port, _err), apptypes.LOG_INFO)
			hm.httpServer = nil
			hm.isstarted = false
			return
		}
	}()
}

// Stop stops the running HttpMonitor
func (hm *HttpMonitor) Stop() {
	if hm.httpServer != nil {
		if _err := hm.httpServer.Shutdown(context.TODO()); _err != nil {
			log.Fatalf("Shutdown(): " +  _err.Error()) // failure/timeout shutting down the server gracefully
		}

		hm.httpServerExitDone.Wait() // wait for goroutine to stop
	}
	hm.isstarted = false
}

func (hm *HttpMonitor) IsStarted() bool {
	return hm.isstarted
}

// setJsonResp sends the JSON message to the client
func (hm *HttpMonitor) setJsonResp(pMessage []byte, pHttpCode int, pResWriter http.ResponseWriter) {
	pResWriter.Header().Set("Content-Type", RESPONSE_FORMAT)
	pResWriter.WriteHeader(pHttpCode)
	pResWriter.Write(pMessage)
}

// live send the liveness of the service
func (hm *HttpMonitor) live(pResWriter http.ResponseWriter, pRequest *http.Request) {
	_status:=struct{Status string}{Status: "LIVE"}
	
	if _message, _err := json.Marshal(_status); _err == nil {
		hm.setJsonResp(_message, http.StatusOK, pResWriter)
		_message=nil
	}
}



// status sends the status message
func (hm *HttpMonitor) status(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if _message, _err := json.Marshal(hm.appInstance.Get_App_Status()); _err == nil {
		hm.setJsonResp(_message, http.StatusOK, pResWriter)
		_message=nil
	}
}

// info sends the information of the application
func (hm *HttpMonitor) info(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if _message, _err := json.Marshal(hm.appInstance.Get_App_Info()); _err == nil {
		hm.setJsonResp(_message, http.StatusOK, pResWriter)
		_message=nil
		
	}
}

// startmonitor starts the web socket monitoring with pre-set address & port
func (hm *HttpMonitor) startmonitor(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if !hm.is_authorized(pRequest) {
		hm.setJsonResp([]byte(""), http.StatusUnauthorized, pResWriter)
		return
	}

	_status:=struct{Status string}{Status: "OK"}
	
	if _, _err := hm.appInstance.Start_WSMonitor(); _err != nil {
		_status.Status="Failed. " + _err.Error()
	} 

	_message, _ := json.Marshal(_status)
	
	defer func ()  {
		_message=nil
	}()
	
	hm.setJsonResp(_message, http.StatusOK, pResWriter)
}

// stopmonitor stops the running web socket monitoring
func (hm *HttpMonitor) stopmonitor(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if !hm.is_authorized(pRequest) {
		hm.setJsonResp([]byte(""), http.StatusUnauthorized, pResWriter)
		return
	}

	fmt.Println("calling Stop_WSMonitor...........")
	_status:=struct{Status string}{Status: "OK"}
	
	if ! hm.appInstance.Stop_WSMonitor() {
		_status.Status="FAILED"
	}
	_message, _ := json.Marshal(_status)
	
	defer func ()  {
		_message=nil
	}()
	
	hm.setJsonResp(_message, http.StatusOK, pResWriter)
	
	fmt.Println("calling Stop_WSMonitor........... DONE")
}

// config_reload reloads the configuration
func (hm *HttpMonitor) config_reload(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if !hm.is_authorized(pRequest) {
		hm.setJsonResp([]byte(""), http.StatusUnauthorized, pResWriter)
		return
	}

	_status:=struct{Status string}{Status: "OK"}

	if _, _err := hm.appInstance.Reload_Config(); _err != nil {
		_status.Status="ERROR. " +  _err.Error()
	}
	_message, _ := json.Marshal(_status)
	
	defer func ()  {
		_message=nil
	}()
	
	hm.setJsonResp(_message, http.StatusOK, pResWriter)
	
}

func (hm *HttpMonitor) set_log_level(pResWriter http.ResponseWriter, pRequest *http.Request) {
	if !hm.is_authorized(pRequest) {
		hm.setJsonResp([]byte(""), http.StatusUnauthorized, pResWriter)
		return
	}
	
	if pRequest.Method != "GET" {
		hm.setJsonResp([]byte(""), http.StatusMethodNotAllowed, pResWriter)
		return
	}
	
	_level := pRequest.URL.Query().Get("level")
	if _level==""{
		http.Error(pResWriter, "invalid log level", http.StatusBadRequest)
		return
	}
	_level=strings.ToLower(_level)
	
	switch _level{
	case "debug":
		hm.appInstance.Set_LogLevel(apptypes.LOG_DEBUG)
	case "info":
		hm.appInstance.Set_LogLevel(apptypes.LOG_INFO)
	case "warn":
		hm.appInstance.Set_LogLevel(apptypes.LOG_WARN)
	case "error":
		hm.appInstance.Set_LogLevel(apptypes.LOG_ERROR)
	case "fatal":
		hm.appInstance.Set_LogLevel(apptypes.LOG_ERROR)
	case "panic":
		hm.appInstance.Set_LogLevel(apptypes.LOG_ERROR)
	}
	
	_status:=struct{Status string}{Status: "OK"}
	
	_message, _ := json.Marshal(_status)
	
	defer func ()  {
		_message=nil
	}()
	
	hm.setJsonResp(_message, http.StatusOK, pResWriter)
	
}

// config_reload reloads the configuration
func (hm *HttpMonitor) config_save(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if !hm.is_authorized(pRequest) {
		hm.setJsonResp([]byte(""), http.StatusUnauthorized, pResWriter)
		return
	}
	
	if pRequest.Method != "POST" {
		hm.setJsonResp([]byte(""), http.StatusMethodNotAllowed, pResWriter)
		return
	}

	_status:=struct{Status string}{Status: "OK"}
	var _message []byte
	
	defer func ()  {
		_message=nil
	}()
	
	_bData, _err := io.ReadAll(pRequest.Body)
	
	defer func ()  {
		_bData=nil
	}()
	
	if _err != nil {
		_status.Status="ERROR. Failed to read the request body. " +  _err.Error()
		_message, _ = json.Marshal(_status)
		hm.setJsonResp(_message, http.StatusNoContent, pResWriter)
		return
	}

	if len(_bData) == 0 {
		_status.Status="ERROR. Invalid content"
		_message, _ = json.Marshal(_status)
		hm.setJsonResp(_message, http.StatusNoContent, pResWriter)
		return
	}
	
	if _, _err := hm.appInstance.Save_App_Config(&_bData); _err != nil {
		
		_status.Status="failed to write content to app configuration file " + _err.Error()
		_message, _ = json.Marshal(_status)
		hm.setJsonResp(_message,http.StatusInternalServerError, pResWriter)
	} else {
		_status.Status="OK"
		_message, _ = json.Marshal(_status)
		hm.setJsonResp(_message, http.StatusOK, pResWriter)
	}
}

//// config_reload reloads the configuration
func (hm *HttpMonitor) list_units(pResWriter http.ResponseWriter, pRequest *http.Request) {
	
	if !hm.is_authorized(pRequest) {
		hm.setJsonResp([]byte(""), http.StatusUnauthorized, pResWriter)
		return
	}
	
	if pRequest.Method != "GET" {
		hm.setJsonResp([]byte(""), http.StatusMethodNotAllowed, pResWriter)
		return
	}
	
	var _units []apptypes.Appunit
	var _err error
	
	defer func ()  {
		_units=nil
		_err=nil
	}()
	
	if _units,_err=hm.appInstance.Units_List();_err!=nil{
		hm.appInstance.Write2Log("error occurred while listing units. " + _err.Error(),apptypes.LOG_ERROR)
		hm.setJsonResp([]byte("error occurred while listing units. please check the error log"), http.StatusInternalServerError, pResWriter)
		return
	}
	
	var _message []byte
	defer func(){
		_message=nil
	}()
	
	if _message,_err=json.Marshal(_units);_err!=nil{
		hm.appInstance.Write2Log("error occurred while parsing units. " + _err.Error(),apptypes.LOG_ERROR)
		hm.setJsonResp([]byte("Invalid unit list. please check the error log"), http.StatusInternalServerError, pResWriter)
		return
	}
	
	hm.setJsonResp(_message, http.StatusOK, pResWriter)
	
}

func (hm *HttpMonitor) stop_unit(pResWriter http.ResponseWriter, pRequest *http.Request) {
	if !hm.is_authorized(pRequest) {
		hm.setJsonResp([]byte(""), http.StatusUnauthorized, pResWriter)
		return
	}
	
	if pRequest.Method != "GET" {
		hm.setJsonResp([]byte(""), http.StatusMethodNotAllowed, pResWriter)
		return
	}
	
	
	_unit_name := pRequest.URL.Query().Get("name")
	if _unit_name==""{
		http.Error(pResWriter, "missing unit name parameter", http.StatusBadRequest)
		return
	}
	
	_force:=pRequest.URL.Query().Get("force")
	
	println("stop uint " + _unit_name + " force=" + _force)
	
	hm.setJsonResp([]byte("TO DD : stop uint " + _unit_name + " force=" + _force), http.StatusOK, pResWriter)

}

func (hm *HttpMonitor) start_unit(pResWriter http.ResponseWriter, pRequest *http.Request) {
	_unit_name := pRequest.URL.Query().Get("name")
	if _unit_name==""{
		http.Error(pResWriter, "missing parameter", http.StatusBadRequest)
	}
	
	println("start uint " + _unit_name)
	hm.setJsonResp([]byte("TO DD : start uint " + _unit_name), http.StatusOK, pResWriter)
}

func (hm *HttpMonitor) restart_unit(pResWriter http.ResponseWriter, pRequest *http.Request) {
	_unit_name := pRequest.URL.Query().Get("name")
	
	if _unit_name==""{
		http.Error(pResWriter, "missing parameter", http.StatusBadRequest)
	}
	
	_force:=pRequest.URL.Query().Get("force")
	
	println("restart uint " + _unit_name + " force=" + _force)
	hm.setJsonResp([]byte("TO DD : restart uint " + _unit_name + " force=" + _force), http.StatusOK, pResWriter)
}

func (hm *HttpMonitor) status_unit(pResWriter http.ResponseWriter, pRequest *http.Request) {
	_unit_name := pRequest.URL.Query().Get("name")
	
	if _unit_name==""{
		http.Error(pResWriter, "missing parameter", http.StatusBadRequest)
	}
	println("send status of uint " + _unit_name)
	
	hm.appInstance.Unit_Status(&_unit_name)
	//for  _unit:=range hm.appInstance.Ap
	hm.setJsonResp([]byte("TO DD : send status of uint " + _unit_name), http.StatusOK, pResWriter)
	
}

var IHTTPMonitor HttpMonitor
