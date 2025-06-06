package issehub

// interface that needs to be implemented for Web Socket monitoring.
type ISSEMonitor interface {

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

	// BroadCastLog broadcasts given log entry to connected log reader client
	Broadcast_Log(message []byte)

	// BroadCastStatus broadcasts given status message among the connected status reader clients
	Broadcast_Status(message []byte)

	// BroadCBroadcast_Monitor_Message broadcasts given message to connected monitor reader clients
	Broadcast_Monitor_Message(message []byte)

	// MonitorsCount returns the number of active monitor message readers count
	Monitor_Clients_Count() int

	// MonitorsCount returns the number of active status readers count
	Status_Clients_Count() int

	// LogClientsCount returns the number of active log readers count
	Log_Clients_Count() int

	// Returns start/stop status of the Web Socket monitoring
	IsStarted() bool
}
