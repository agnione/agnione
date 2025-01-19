// main package provides launcher for Kandy application framework
//
//  - Creates an instance of Agni Application Framework
//	- Initialize and starts the Agni
//	- Watch OS Intercept signals and stops the running Agni instance
//
// This package includes functions:
//	- BuildInfo
//	- GetBasePath
//	- SignalHandler
//  - usage
//	- main
/*
#########################################################################################

	Author        :  D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024

	Copyright     :  © 2024 D. Ajith Nilantha de Silva contact@agnione.net
						Licensed under the Apache License, Version 2.0 (the "License");
						you may not use this file except in compliance with the License.
						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0

						Unless required by applicable law or agreed to in writing, software
						distributed under the License is distributed on an "AS IS" BASIS,
						WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
						See the License for the specific language governing permissions and
						limitations under the License.

	Class/module  :  app

	Objective     :  Provide the application launcher/shell to start the KAndy Application Framework
	 		  			main entry of the KAF
#######################################################################################################################

	Author                 	Date        	Action      	Description
	--------------------------------------------------------------------------------------------------------------------

	Ajith de Silva		24/10/2023	Created 	Created the initial version

	Ajith de Silva		29/10/2023	Updated 	Added the interrupt handler

	Ajith de Silva		12/11/2024	Updated 	Added the functions

	Ajith de Silva		06/03/2024	Updated 	Added the log path as command line argument

#######################################################################################################################
*/
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	build "agnione/v1/src/lib" /// import the AgniOne lib package

	agni "agnione.appfm/src/core" /// import the AgniOne application framework packages
	"github.com/fatih/color"

	libbuild "agnione.appfm/src/build"
)


var agniApp *agni.AgniApp /// global Agni application instance 
var cancelFunc context.CancelFunc
var ctx context.Context

// struct to hold application build information
func BuildInfo() *build.BuildInfo {
	return &build.BuildInfo{
		BuildGoVersion: libbuild.BuildGoVersion,
		Time:           libbuild.Time,
		User:           libbuild.User, Version: libbuild.Version}
}

// GetBasePath returns application base bath
func GetBasePath() *string {

	var _curDir, _err = os.Getwd()
	if _err != nil {
		_curDir, _ = os.Executable()
		_curDir = filepath.Dir(_curDir)  + "/"
	}
	
	return &_curDir
}

// signaleHandler wait for the interrupt signal to support graceful shutdown.
func SignaleHandler() {
	<-ctx.Done()
	println("*********************************\nShutdown Signal Received\n*********************************")
}

/// define the command line arguments
var main_path = flag.String("main_path", "", "base/root path of the application")
var log_path = flag.String("log_path", "", "path that application writes the log entries. if not given, application will use path in config file")
var app_path = flag.String("app_path", "", "base path that application configuration file (app.config) exists. If not given then, will try to load app.config from <main_path>")
var cpu_count = flag.Int("cpu_count", 5, "number of cpu cores to be used. If not given, all available cpu cores will be used.")
var rest_port= flag.Int("rest_port", 8080, "TCP port that application exposes its REST endpoints to control & monitor application. default it 8080. Max:65635.")
var ws_port=flag.Int("ws_port", 2345, "TCP port that application exposes its web socket endpoints for real time application monitor. Default it 2345. Max:65635.")

func Filter_Number(value string) int {
	if _value,_err:=strconv.Atoi(value); _err != nil {
		return 0
	}else{
		return _value
	}
}

func usage() {
	println("usage: app --main_path <app_base_path> --log_path <app_log_path> --app_path <app_config_path>  --restport <8880>  --wsport <23450> --cpucount 4")
	flag.PrintDefaults()
	println("\n** if main_path is not given then application will use the '<executable_folder>' as main_path by default.")
	println("** if log_path is not given then application will use the pre-set paths in config file")
	println("** if app_path is not given then application will use the <main_path>as app.config path")
}

///// main entry point
func main() {

	defer func ()  {
		if _r:=recover();_r!=nil{
			fmt.Printf("Recovered panic %v",_r)
			_r=nil
		}
		runtime.GC()
	}()

	println("");println("")
	color.Cyan(banner)
	banner = ""
	println("");println("")

	_buildinfo := BuildInfo() /// get the application build information
	color.Yellow("\tVersion : " + _buildinfo.Version + "\n\tBuilt time : " +  _buildinfo.Time + "\n\tBuilt user : " + _buildinfo.User + "\n\tBuilt Go version : " + _buildinfo.BuildGoVersion + "\n\n")
	_buildinfo = nil
	color.Cyan("############################################################\n\n")

	flag.Usage = usage
	
	// if help is passed as cmd line. Show usage and exit.
	if len(os.Args) == 2 && strings.ToLower(os.Args[1]) == "--help" {
		flag.Usage()
		return
	}

	flag.Parse() /// parse the arguments

	/// if main app config path not set. then use the default relative path
	if *main_path == "" {
		main_path = GetBasePath()
	}

	/// if unit_path not set then set main_path as config_path as default
	if *app_path == "" {
		app_path = main_path
	}

	defer func() {

		if _r:=recover();_r!=nil{
			fmt.Printf("Recovered panic %v",_r)
			_r=nil
		}
		
		main_path=nil
		log_path=nil
		app_path=nil
		
		rest_port=nil
		ws_port=nil
		cpu_count=nil
	}()

	/// check the ports values in cmd line
	/// set default values to 0
	var _cpu_count int=runtime.NumCPU()	//// set teh default CPU cores 
	var _os_pid=os.Getpid()
	
	/// check for CPU count validity
	if *cpu_count==0{
		cpu_count=&_cpu_count
	}
	
	if *cpu_count>_cpu_count{
		cpu_count=&_cpu_count
	}
	
	runtime.GOMAXPROCS(*cpu_count)	/// set max CPU for go runtime
	
	println("CPU cores     : " +  strconv.Itoa(*cpu_count) + "/" + strconv.Itoa(runtime.NumCPU()))
	println("OS Process ID : " + strconv.Itoa(_os_pid))
	
	var _err error

	/// read config from config server and save it to the /config folder
start:

	/// create AgniOne App instance and initialize it
	println("using app root path\t: " +  *main_path)
	println("using app log path\t: "  +  *log_path)
	println("using app unit path\t: " +  *app_path + "\n")
	
	println("\nInitialzing AgniOne ......")
	agniApp = new(agni.AgniApp)

	/// create the cancellation application context
	ctx, cancelFunc = context.WithCancel(context.Background())
	
	defer func ()  {
		ctx=nil
		cancelFunc=nil
		agniApp=nil
	}()
	
	/// initialize the app framework using the parameters.
	if _, _err = agniApp.Initialize(&ctx, &_os_pid, main_path,  app_path, log_path,rest_port,ws_port); _err != nil {
		println("error " + _err.Error() + ". AgniOne is terminating")
		return
	}
	
	println("Initializing AgniOne ............  DONE")
	
	/* SIGNAL handling section */
	termChan := make(chan os.Signal, 1) // Handle sigterm and await terminate signal CTRL + C signal
	signal.Notify(termChan, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGABRT)
	
	defer func ()  {
		termChan=nil
	}()
	
	go SignaleHandler() /// starts the interrupt signal handler

	println("Starting ::" + agniApp.Name() + " (" + agniApp.ID() + ") - " + agniApp.Version())

	agniApp.Start() /// starts the application framework

	println("Started :: " + agniApp.Name() + " (" + agniApp.ID() + ") - " + agniApp.Version())
	
	<-termChan // Blocks here until interrupt occur

	cancelFunc() /// initiate the context Cancel

	println("Signalling the AgniOne Unit(s) to stop")
	agniApp.Stop()                 /// call the stop method of the AgniOne Framework. This will trigger all routines in it to stop
	time.Sleep(time.Second * 5) /// give some time to stop/cleanup routines

	println("Flag all routines to stop.......")
	agniApp.Terminate() // broadcast the main channel stopped message
	println("Flag all routines to stop....... DONE")

	println("Waiting for termination of routines.......")
	agniApp.WaitforClose() // Wait until all routines are done
	println("All routines terminated")
	time.Sleep(time.Second * 1)
	reload := agniApp.Reload_Requested() /// read if reload requested flag set

	/// clear the variables
	println("Stopped :: " + agniApp.Name() + " (" + agniApp.ID() + ") - " + agniApp.Version() + "\n")
	println("Cleaning the Agni environment")
	agniApp.DeInitialize()	///
	agniApp = nil

	/// if reload requested then reload the Agni
	if reload {
		println("Application reload requested.\r\n Reloading application....")
		runtime.GC()
		goto start
	}

	ctx = nil
	termChan = nil

	println("\nAgniOne Framework terminated.\n")
}

var banner = `
############################################################

▄▄▄        ▄████  ███▄    █  ██▓
▒████▄     ██▒ ▀█▒ ██ ▀█   █ ▓██▒
▒██  ▀█▄  ▒██░▄▄▄░▓██  ▀█ ██▒▒██▒
░██▄▄▄▄██ ░▓█  ██▓▓██▒  ▐▌██▒░██░
 ▓█   ▓██▒░▒▓███▀▒▒██░   ▓██░░██░
 ▒▒   ▓▒█░ ░▒   ▒ ░ ▒░   ▒ ▒ ░▓  
  ▒   ▒▒ ░  ░   ░ ░ ░░   ░ ▒░ ▒ ░
  ░   ▒   ░ ░   ░    ░   ░ ░  ▒ ░
      ░  ░      ░          ░  ░  
                                 
###############   AgniOne Application Framework ##############
Designed & Developed by D. Ajith Nilantha de Silva
© 2025 D. Ajith Nilantha de Silva contact@agnione.net
############################################################
`