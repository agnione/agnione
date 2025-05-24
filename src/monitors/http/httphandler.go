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
	iappfw "agnione/v2/src/appfm/iappfw" /// import interface of AgniOne
	apptypes "agnione/v2/src/appfm/types"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"agnione.appfm/src/monitors/sse/ssehandler"
)

// RESPONSE_FORMAT define the default response format to application/JSON
const RESPONSE_FORMAT = "application/json;charset=utf-8"

// HttpMonitor struct of the HttpMonitor
type HttpMonitor struct {
	appInstance        iappfw.IAgniApp
	httpServerExitDone *sync.WaitGroup
	sseMonitor         *ssehandler.SSEMonitor
	apikeys            *[]string
	cors               *[]string
	httpServer         *http.Server
	httpAddr           string
	isstarted          bool
}

//go:embed ui/*.html
var embededUIs embed.FS

// Initialize initializes the HttpMonitor instance
// Requires the IZApp interface as parameter
func (hm *HttpMonitor) Initialize(pApp_Instance iappfw.IAgniApp) {
	hm.appInstance = pApp_Instance
	hm.httpServerExitDone = &sync.WaitGroup{}

	var _err error

	var _temp_path = *hm.appInstance.App_Path() + "config/apikeys.config"
	/// load the api keys to REST authentication
	hm.apikeys, _err = hm.appInstance.Get_FileContent_Lines(&_temp_path)
	if _err != nil {
		hm.appInstance.Write2Log("HTTP Monitor :: failed to read the "+_temp_path+" - "+_err.Error(), apptypes.LOG_ERROR)
	}

	_temp_path = *hm.appInstance.App_Path() + "config/cores.config"
	/// load the api keys to REST authentication
	hm.cors, _err = hm.appInstance.Get_FileContent_Lines(&_temp_path)
	if _err != nil {
		hm.appInstance.Write2Log("HTTP Monitor :: failed to read CORS entries. "+_temp_path+" - "+_err.Error(), apptypes.LOG_ERROR)
		hm.cors = &[]string{"*"} /// set Allow all by default
	}

	/// creates the instance of SSEMonitor & initialize it
	hm.sseMonitor = &ssehandler.SSEMonitor{}

	hm.sseMonitor.Initialize(log.New(os.Stdout, "http: ", log.LstdFlags))
	_temp_path = ""
	hm.isstarted = false

}

func (hm *HttpMonitor) DeInitialize() {
	hm.appInstance = nil
	hm.apikeys = nil
	hm.httpServerExitDone = nil
	hm.httpServer = nil

	if hm.sseMonitor != nil {
		hm.sseMonitor.DeInitialize()
		hm.sseMonitor = nil
	}
}

func (hm *HttpMonitor) Get_SSEMonitorPtr() *ssehandler.SSEMonitor {
	return hm.sseMonitor
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

func (hm *HttpMonitor) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Get the Origin header from the incoming request
		_origin := r.Header.Get("Origin")
		_ok := false

		// Check if the origin is in the allowedOrigins list
		for _, _allowed := range *hm.cors {
			if _ok = strings.EqualFold(_origin, _allowed); _ok {
				break
			}
		}
		if !_ok {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", _origin)
		w.Header().Add("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		w.Header().Add("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, apikey, accept, origin, Cache-Control, X-Requested-With")

		next.ServeHTTP(w, r)

	})
}

// is_authorized checks if the apikey is given in the HTTP request header
//
// if given apikey is valid then returns true
// Unless returns false
func (hm *HttpMonitor) is_authorized(pRequest *http.Request) bool {

	_apiKey := pRequest.Header.Get("apikey")
	if _apiKey == "" {
		_apiKey = pRequest.URL.Query().Get("auth")
	}
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

func (hm *HttpMonitor) Render_View_Handler(w http.ResponseWriter, pViewTemplate string, params any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if _tmpl, _err := template.ParseFS(embededUIs, pViewTemplate); _err != nil {
		hm.appInstance.Write2Log("Failed to parse embeded "+pViewTemplate+". "+_err.Error(), apptypes.LOG_ERROR)
	} else {
		_tmpl.Execute(w, params)
	}
}

func (hm *HttpMonitor) Log_View_Handler(w http.ResponseWriter, r *http.Request) {

	hm.Render_View_Handler(w, "ui/monitor_client.html", map[string]string{
		"Name":     "Log",
		"Endpoint": "log",
		"Addr":     hm.httpAddr,
	})
}

func (hm *HttpMonitor) Status_View_Handler(w http.ResponseWriter, r *http.Request) {

	hm.Render_View_Handler(w, "ui/monitor_client.html", map[string]string{
		"Name":     "Status",
		"Endpoint": "status",
		"Addr":     hm.httpAddr,
	})
}

func (hm *HttpMonitor) Event_View_Handler(w http.ResponseWriter, r *http.Request) {

	hm.Render_View_Handler(w, "ui/monitor_client.html", map[string]string{
		"Name":     "Event",
		"Endpoint": "events",
		"Addr":     hm.httpAddr,
	})
}

// Start starts the HttpMonitor on given address and port
// Parameter address is the host address
// Parameter http_port is the port to bind on given address
func (hm *HttpMonitor) Start(pAddress *string, pHttp_Port *int) {

	hm.httpServerExitDone.Add(1)

	hm.httpAddr = fmt.Sprintf("%s:%d", *pAddress, *pHttp_Port)

	go func() {

		/// check if already started
		if hm.isstarted {
			return
		}

		defer func() {
			if _r := recover(); _r != nil {
				fmt.Println("Recovered panic ", _r)
				_r = nil
			}

			hm.httpServer = nil
			hm.httpServerExitDone.Done()
			hm.isstarted = false
		}()

		_mux := http.NewServeMux()

		_mux.HandleFunc("/live", hm.live)

		_mux.Handle("/ready", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.ready))))

		///apikey authentication have been set to below routes
		_mux.Handle("/info", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.info))))
		_mux.Handle("/status", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.status))))

		/// configuration management

		_mux.Handle("/admin/config/reload", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.config_reload))))
		_mux.Handle("/admin/config/save", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.config_save))))
		_mux.Handle("/admin/log/setlevel", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.set_log_level))))

		/// application units management

		_mux.Handle("/admin/units", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.list_units))))
		_mux.Handle("/admin/unit/stop", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.stop_unit))))
		_mux.Handle("/admin/unit/{name}/start", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.start_unit))))
		_mux.Handle("/admin/unit/{name}/restart?force=", hm.authMiddleware(http.HandlerFunc(hm.restart_unit)))
		_mux.Handle("/admin/unit/{name}/status", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.status_unit))))

		//// set teh SSE monitoring
		if hm.sseMonitor != nil {

			_mux.Handle("/monitor/view/logs", http.HandlerFunc(hm.Log_View_Handler))
			_mux.Handle("/monitor/view/status", http.HandlerFunc(hm.Status_View_Handler))
			_mux.Handle("/monitor/view/events", http.HandlerFunc(hm.Event_View_Handler))

			_mux.Handle("/monitor/log/read", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.sseMonitor.Logger))))
			_mux.Handle("/monitor/status/read", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.sseMonitor.Status))))
			_mux.Handle("/monitor/events/read", hm.corsMiddleware(hm.authMiddleware(http.HandlerFunc(hm.sseMonitor.Monitor))))

			hm.sseMonitor.Start()
		}

		hm.isstarted = true

		/// need to set origins via config
		hm.httpServer = &http.Server{Addr: hm.httpAddr,
			DisableGeneralOptionsHandler: true, Handler: _mux,
		}

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

	defer func() {
		if _r := recover(); _r != nil {
			fmt.Println("Recovered panic from HTTP Monitor STOP. ", _r)
			_r = nil
		}
	}()

	if hm.httpServer != nil {
		shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownRelease()

		if _err := hm.httpServer.Shutdown(shutdownCtx); _err != nil {
			hm.appInstance.Write2Log("HTTP Server Shutdown(): "+_err.Error(), apptypes.LOG_FATAL) // failure/timeout shutting down the server gracefully
		}
		hm.httpServerExitDone.Wait() // wait for goroutine to stop

		hm.isstarted = false
	}

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
	_status := struct{ Status string }{Status: "LIVE"}

	if _message, _err := json.Marshal(_status); _err == nil {
		hm.setJsonResp(_message, http.StatusOK, pResWriter)
		_message = nil
	}
}

// status sends the status message
func (hm *HttpMonitor) status(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if _message, _err := json.Marshal(hm.appInstance.Get_App_Status()); _err == nil {
		hm.setJsonResp(_message, http.StatusOK, pResWriter)
		_message = nil
	} else {
		hm.appInstance.Write2Log("API Error occured while reading Status. "+_err.Error(), apptypes.LOG_ERROR)
	}
}

// info sends the information of the application
func (hm *HttpMonitor) info(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if _message, _err := json.Marshal(hm.appInstance.Get_App_Info()); _err == nil {
		hm.setJsonResp(_message, http.StatusOK, pResWriter)
		_message = nil
	} else {
		hm.appInstance.Write2Log("API Error occured while reading info. "+_err.Error(), apptypes.LOG_ERROR)
	}
}

// info sends the information of the application
func (hm *HttpMonitor) ready(pResWriter http.ResponseWriter, pRequest *http.Request) {

	///1. get the loaded Agni Unites Ready status
	///2. return response

	if _message, _err := json.Marshal(hm.appInstance.Get_Ready_Status()); _err == nil {
		hm.setJsonResp(_message, http.StatusOK, pResWriter)
		_message = nil
	} else {
		hm.appInstance.Write2Log("API Error occured while reading Ready Status. "+_err.Error(), apptypes.LOG_ERROR)
	}
}

// config_reload reloads the configuration
func (hm *HttpMonitor) config_reload(pResWriter http.ResponseWriter, pRequest *http.Request) {

	_status := struct{ Status string }{Status: "OK"}

	if _, _err := hm.appInstance.Reload_Config(); _err != nil {
		_status.Status = "ERROR. " + _err.Error()
	}
	_message, _ := json.Marshal(_status)

	defer func() {
		_message = nil
	}()

	hm.setJsonResp(_message, http.StatusOK, pResWriter)

}

func (hm *HttpMonitor) set_log_level(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if pRequest.Method != "GET" {
		hm.setJsonResp([]byte(""), http.StatusMethodNotAllowed, pResWriter)
		return
	}

	_level := pRequest.URL.Query().Get("level")
	if _level == "" {
		http.Error(pResWriter, "invalid log level", http.StatusBadRequest)
		return
	}
	_level = strings.ToLower(_level)

	switch _level {
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

	_message, _ := json.Marshal(struct{ Status string }{Status: "OK"})

	defer func() {
		_message = nil
	}()

	hm.setJsonResp(_message, http.StatusOK, pResWriter)

}

// config_reload reloads the configuration
func (hm *HttpMonitor) config_save(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if pRequest.Method != "POST" {
		hm.setJsonResp([]byte(""), http.StatusMethodNotAllowed, pResWriter)
		return
	}

	_status := struct{ Status string }{Status: "OK"}
	var _message []byte

	defer func() {
		_message = nil
	}()

	_bData, _err := io.ReadAll(pRequest.Body)

	defer func() {
		_bData = nil
		_err = nil
	}()

	if _err != nil {
		_status.Status = "ERROR. Failed to read the request body. " + _err.Error()
		_message, _ = json.Marshal(_status)
		hm.setJsonResp(_message, http.StatusNoContent, pResWriter)
		return
	}

	if len(_bData) == 0 {
		_status.Status = "ERROR. Invalid content"
		_message, _ = json.Marshal(_status)
		hm.setJsonResp(_message, http.StatusNoContent, pResWriter)
		return
	}

	if _, _err := hm.appInstance.Save_App_Config(&_bData); _err != nil {

		_status.Status = "failed to write content to app configuration file " + _err.Error()
		_message, _ = json.Marshal(_status)
		hm.setJsonResp(_message, http.StatusInternalServerError, pResWriter)
	} else {
		_status.Status = "OK"
		_message, _ = json.Marshal(_status)
		hm.setJsonResp(_message, http.StatusOK, pResWriter)
	}
}

// // config_reload reloads the configuration
func (hm *HttpMonitor) list_units(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if pRequest.Method != "GET" {
		hm.setJsonResp([]byte(""), http.StatusMethodNotAllowed, pResWriter)
		return
	}

	var _units []apptypes.Appunit
	var _err error

	defer func() {
		_units = nil
		_err = nil
	}()

	if _units, _err = hm.appInstance.Units_List(); _err != nil {
		hm.appInstance.Write2Log("error occurred while listing units. "+_err.Error(), apptypes.LOG_ERROR)
		hm.setJsonResp([]byte("error occurred while listing units. please check the error log"), http.StatusInternalServerError, pResWriter)
		return
	}

	var _message []byte
	defer func() {
		_message = nil
	}()

	if _message, _err = json.Marshal(_units); _err != nil {
		hm.appInstance.Write2Log("error occurred while parsing units. "+_err.Error(), apptypes.LOG_ERROR)
		hm.setJsonResp([]byte("Invalid unit list. please check the error log"), http.StatusInternalServerError, pResWriter)
		return
	}

	hm.setJsonResp(_message, http.StatusOK, pResWriter)

}

func (hm *HttpMonitor) stop_unit(pResWriter http.ResponseWriter, pRequest *http.Request) {

	if pRequest.Method != "GET" {
		hm.setJsonResp([]byte(""), http.StatusMethodNotAllowed, pResWriter)
		return
	}

	_unit_name := pRequest.URL.Query().Get("name")
	if _unit_name == "" {
		http.Error(pResWriter, "missing unit name parameter", http.StatusBadRequest)
		return
	}

	_bforce := false
	if pRequest.URL.Query().Get("force") == "true" {
		_bforce = true
	}

	if _, _err := hm.appInstance.Unit_Stop(&_unit_name, _bforce); _err != nil {
		hm.setJsonResp([]byte("Failed to stop unit "+_unit_name+". error "+_err.Error()), http.StatusOK, pResWriter)
	} else {
		hm.setJsonResp([]byte("Unit stop "+_unit_name), http.StatusOK, pResWriter)
	}
}

func (hm *HttpMonitor) start_unit(pResWriter http.ResponseWriter, pRequest *http.Request) {
	_unit_name := pRequest.URL.Query().Get("name")
	if _unit_name == "" {
		http.Error(pResWriter, "missing parameter", http.StatusBadRequest)
	}

	println("start uint " + _unit_name)

	if _, _err := hm.appInstance.Unit_Start(&_unit_name); _err != nil {
		hm.setJsonResp([]byte("Failed to start unit "+_unit_name+". error "+_err.Error()), http.StatusOK, pResWriter)
	} else {
		hm.setJsonResp([]byte("Unit started "+_unit_name), http.StatusOK, pResWriter)
	}

}

func (hm *HttpMonitor) restart_unit(pResWriter http.ResponseWriter, pRequest *http.Request) {
	_unit_name := pRequest.URL.Query().Get("name")

	if _unit_name == "" {
		http.Error(pResWriter, "missing parameter", http.StatusBadRequest)
	}

	_bforce := false
	if pRequest.URL.Query().Get("force") == "true" {
		_bforce = true
	}

	if _, _err := hm.appInstance.Unit_Restart(&_unit_name, _bforce); _err != nil {
		hm.setJsonResp([]byte("Failed to Re-Start unit "+_unit_name+". error "+_err.Error()), http.StatusOK, pResWriter)
	} else {
		hm.setJsonResp([]byte("Unit Re-Started "+_unit_name), http.StatusOK, pResWriter)
	}
}

func (hm *HttpMonitor) status_unit(pResWriter http.ResponseWriter, pRequest *http.Request) {
	_unit_name := pRequest.URL.Query().Get("name")

	if _unit_name == "" {
		http.Error(pResWriter, "missing parameter", http.StatusBadRequest)
	}

	hm.appInstance.Unit_Status(&_unit_name)
	//for  _unit:=range hm.appInstance.Ap
	hm.setJsonResp([]byte("TO DD : send status of uint "+_unit_name), http.StatusOK, pResWriter)

}

var IHTTPMonitor HttpMonitor
