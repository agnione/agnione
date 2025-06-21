#! /bin/sh


## set defult paths
MAINPATH="/home/agnione/"
APPPATH="/home/agnione/apps/configs/test.config"  ### set the default Unit config file. this should mount with docker HOST
LOGPATH="/home/agnione/logs/" ### set the default log path. this should mount with docker HOST

cd /home/agnione

### run the ZAF with params
./agnione.app --mainpath $MAINPATH --apppath $APPPATH --logpath $LOGPATH --restport 0 --cpu_count 0
