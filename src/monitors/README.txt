all the monitoring packages are here

REST endpoints for Application
1. info
    http://localhost:8080/info

2. status    
    http://localhost:8080/status

3. Liveness
    http://localhost:8080/live

4. view loat 100 lines of log
    http://localhost:8080/admin/log

5. reload configuration
    http://localhost:8080/admin/config/reload		

6. stop websocket monitoring
    http://localhost:8080/admin/monitor/stop

7. start websocket monitoring
    http://localhost:8080/admin/monitor/start




## monitoring UI for WS
1. status monitor
    http://localhost:2345/wsstatus

2. realtime monitor
    http://localhost:2345/wsmonitor


3. log monitor
    http://localhost:2345/wslogger

## websocket endpoints

ws://localhost:2345/app/status  -- ws status end point
ws://localhost:2345/app/monitor -- ws monitor realtime activity of the application
ws://localhost:2345/app/logger -- ws log monitor

## profiler
    Helpes to cllect CPU and memory profle data on-demand.

    1. start the profiler
        http://localhost:8080/admin/profiler/start

    2.  stop profiler
        http://localhost:8080/admin/profiler/stop


