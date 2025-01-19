package ihttpmonitor

import agniapp "agnione/v1/src/appfm/iappfw"
type IHTTPMonitor interface {
	Initialize(agniapp.IAgniApp)
	DeInitialize()
	Start(pAddress string, pHttp_Port int8)
	Stop()
	IsStarted() bool
}
