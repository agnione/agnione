# Agni Application Framework - Monitoring/Controlling

## About
Agni application framework is a generic High Performance extendable modular system written in Go (https://go.dev/) for Unix based systems.<br>

When the application is running with N number of units it is possible to monitor its activities by its log files.
In order to do that person who monitor has to be in the same machine or use any observability tool so that it can be monitored remotely.

AgnoOne Application Framework is having build-in REST and Wsb Socket monitoring features.

REST Monitoring is on by default and Web Socket Monitoring had to be switched on via REST API calls.
<br/> <br/>

![]()<img src="./asserts/icon_image.png" width="150px" >
## REST Monitoring

  There 3 types of REST APIs
   
  1. Retrieve Information
  2. Control Application
  3. Control Web Socket Monitoring
   
   
  All the HTTP REST endpoint will be hosted at http://localhost:8080 by default.
   
  Check liveness -> http://localhost:8080/live

  Rest of he end points are expecting HTTP header "apikey" with valid key which is given in the AgniOne config/apikeys.config

 #### it is possible to set the log level at any time using 
  http://localhost:8080/admin/log/setlevel?level=<LOG_LEVEL>
  <br/>valid prams are <b>info,warn,debug,error </b>

eg:- <br/>
  http://localhost:8080/admin/log/setlevel?level=info
  http://localhost:8080/admin/log/setlevel?level=warn
  
  <br/> <br/>

![]()<img src="./asserts/websocket_client.png" width="150px" >
### Web Socket Monitring

In order to monitor the application real-time activities over web socket, it is required to start the AgniOne built-in web socket monitoring feature.
#### start web sokcet monitorin
  URL: http://localhost:8080/admin/monitor/start
  METHOD: GET
  HTTP-HEADER apikey:09E64D1428F9854F16DBBEEC1AFA6270
  
  When the web socket monitor started there are 3 real-time monitoring facilities.
  1. Real-time status monitor -> http://localhost:2345/wsstatus
  2. Real-time log monitor -> http://localhost:2345/wslogger
  3. Real-time monitor message viewer -> http://localhost:2345/wsmonitor

Also these web socket endpoints can be hooked up to external monitoring applications and act based on the received information.

When monitoring over web socket is done, it is recommended to shutdown it using
 URL http://localhost:8080/admin/monitor/start with valid apikey.
 
 

