// build package provides Interface to implement the Logger plugin for AgniOne Application Framework
//
// This interfce defines functions that needs to implement when building ZFP Logger plugin
//
// - Initialize
// - GetID
// - DeInitialize
// - WriteDebug
// - WriteWarn
// - WriteInfo
// - WriteError
// - WriteFatal
// - BuildInfo
/*
#########################################################################################
Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024

Copyright     :   contact@agnione.net

Class/module  :   alogger.go

Objective     :   Define the logger client interface plugin

	This package has ability to handle all the logging levels.
	It is required to provide ialogger/ialogger.go and klogger/klogger.go
	files to build the client plug-in.

#########################################################################################

	Author			Date		Action		Description

#########################################################################################

Ajith de Silva				  11/01/2024  Optimized		Optimized the variables and channel
Ajith de Silva				  16/12/2024  Updated		Added set_loge_level method

#########################################################################################
*/
package logger

import (
	atypes "agnione/v1/src/appfm/types"
	"fmt"

	iappfw "agnione/v1/src/appfm/iappfw" /// import interface of Agni

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

type ALogger struct {	
	logmessage chan LogMessage
	appInstace iappfw.IAgniApp
	logger     zerolog.Logger
	ljkLogger *lumberjack.Logger
	stopper    chan bool
	id         int
	IS_Started bool
	is_initialized bool
}

type LogMessage struct {
	Msg_Entry string
	Msg_Type  atypes.LogLevel
}

// Initialize : Initialize the instance with filepath, loglevel and instance_id parameters.
// Returns true if the file path is exists
// Returns false if the file path isn't exist
func (al *ALogger) Initialize(app_instance iappfw.IAgniApp, filepath string, log_level atypes.LogLevel, instance_id int) (bool, error) {
	
	if app_instance == nil {
		return false, fmt.Errorf("application instance is not initialized")
	}
	
	al.appInstace = app_instance

	al.ljkLogger = &lumberjack.Logger{
        Filename: filepath,
        MaxSize: 5,
      	MaxAge:     28,
      	Compress:   true,
		LocalTime: true,
	  
    }

	al.logger =zerolog.New(al.ljkLogger).With().Timestamp().Logger()
	
	al.id = instance_id
	al.logger.Level(al.get_log_level(log_level))
	al.logmessage = make(chan LogMessage)
	
	return true, nil
}

// GetID returns the pre-set id of the current instance
func (al *ALogger) GetID() (instance_id int) {
	return al.id
}

func (al *ALogger) get_log_level(log_level atypes.LogLevel) zerolog.Level  {
	
	switch log_level {
	
	case atypes.LOG_DEBUG:
		return zerolog.DebugLevel
	case atypes.LOG_INFO:
		return zerolog.InfoLevel
	case atypes.LOG_WARN:
		return zerolog.WarnLevel
	case atypes.LOG_ERROR:
			return zerolog.ErrorLevel
	case atypes.LOG_PANIC:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}


// Clear resources and stop the channel
func (al *ALogger) DeInitialize() {
	close(al.logmessage)
	al.appInstace = nil
	
	al.ljkLogger=nil
}


func (al *ALogger) Start() bool {

	al.stopper = make(chan bool)
	go al.log_writer()

	al.IS_Started = true
	al.is_initialized =true
	al.logger.Info().Msg("Logger Initialized & Started")
	return al.IS_Started
}

func (al *ALogger) Stop() bool {
	
	defer recover()
	
	if !al.IS_Started {
		return al.IS_Started
	}
	close(al.stopper)
	
	al.IS_Started = false
	al.ljkLogger.Close()
	return al.IS_Started
}

func (al *ALogger) Set_LogLevel(log_level atypes.LogLevel) {
	al.logger.Level(al.get_log_level(log_level))
}


// log_writer writes log entries according to the message type
func (al *ALogger) log_writer() {

	var _mlogmsg LogMessage
	_ok:=false
	
	defer func ()  {
		_mlogmsg=LogMessage{}
	}()
	for {
		select {
			case <-al.stopper:
				return

			case _mlogmsg,_ok = <-al.logmessage:{
				
				if !_ok || ! al.is_initialized  {
					return
				}

				// Writes log according to the log level
				switch _mlogmsg.Msg_Type {
					case atypes.LOG_ERROR:
						al.logger.Error().Msg(_mlogmsg.Msg_Entry)
					case atypes.LOG_WARN:
						al.logger.Warn().Msg(_mlogmsg.Msg_Entry)
					case atypes.LOG_INFO:
						al.logger.Info().Msg(_mlogmsg.Msg_Entry)
					case atypes.LOG_DEBUG:
						al.logger.Debug().Msg(_mlogmsg.Msg_Entry)
					case atypes.LOG_PANIC:
						al.logger.Error().Msg("**PANIC**" + _mlogmsg.Msg_Entry)
				}
			}
		}

	}
}

// WriteDebug Writes the log entry in debug level
func (al *ALogger) WriteLog(pLogMessage LogMessage) {
	
	if al.IS_Started {
		al.logmessage <- pLogMessage
	}
}

// WriteDebug Writes the log entry in debug level
func (al *ALogger) WriteDebug(pEntry string) {
	al.logmessage <- LogMessage{Msg_Entry: pEntry, Msg_Type: atypes.LOG_DEBUG}
}

// WriteWarn Writes the log entry in warning level
func (al *ALogger) WriteWarn(pEntry string) {
	al.logmessage <- LogMessage{Msg_Entry: pEntry, Msg_Type: atypes.LOG_WARN}
}

// WriteInfo Writes the log entry in information level
func (al *ALogger) WriteInfo(pEntry string) {
	al.logmessage <- LogMessage{Msg_Entry: pEntry, Msg_Type: atypes.LOG_INFO}
}

// WriteError Writes the log entry in error level
func (al *ALogger) WriteError(pEntry string) {
	al.logmessage <- LogMessage{Msg_Entry: pEntry, Msg_Type: atypes.LOG_ERROR}
}

// WriteFatal Writes the log entry in fatal level
func (al *ALogger) WriteFatal(pEntry string) {
	al.logmessage <- LogMessage{Msg_Entry: pEntry, Msg_Type: atypes.LOG_FATAL}
}

// WritePanic Writes the log entry in error level
func (al *ALogger) WritePanic(pEntry string) {
	/// Type "Error" is used here to prevent the process from stalling
	al.logmessage <- LogMessage{Msg_Entry: pEntry, Msg_Type: atypes.LOG_PANIC} 
}
