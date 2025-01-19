![]()<img src="./asserts/Logo_1_transparent.png" >
# AgniOne Application Framework .V1


## Getting started
Please make sure that Go (https://go.dev/) install and configured on you development PC.<br>
If not, please refer https://go.dev/doc/install 

### Get the source code & build step by step

1. Prepare parent folder
   
 ```
  mkdir ~/AgniOneFM
  cd ~/AgniOneFM
   ```
   
2. Clone the AgniOne Framework package to <$GOROOT>/src folder

```
cd ~/AgniOneFM

git clone git@github.com:agnione/libs.git
cd ./libs/v1

chmod 775 ./update_local.sh
./update_local.sh
```
*** ./update_local.sh script will copy AgniOne packages into $GOROOT/src folder
If not, please check your $GOROOT env variable is set.


3. Clone the AgniOne PlugIns to your plugins folder
```
cd ~/AgniOneFM/
git clone git@github.com:agnione/plugins.git
```
3.1 Build the plugin for errors.

```
cd ./plugins

#set execution permission for all the bash scripts
find . -name "*.sh" -exec chmod 775 {} \;

## build script require the destination path to copy the binaries
./build-plugins.sh ~/AgniOne/agnione
```
<b>Should any errors relates to AgniOne packages means the AgniOne packages are not located in the $GOROOT path. Please verify.</b>
*** It should be located at $GOROOT/src/agnione/v1

4. Clone the AgniOne Application Framework to your project folder
```
cd ~/AgniOneFM
git clone git@github.com:agnione/agnione.git
cd ./agniOne
```

4.1 Build AgniOne Framework
```
#set execution permission for all the bash scripts
find . -name "*.sh" -exec chmod 775 {} \;

#build the AgniOne Framework
./build.sh

```
5. Clone the Demo HTTP AgniOne Units your development unit folder

```
cd ~/AgniOneFM
  
git clone git@github.com:agnione/units.git
cd units
```

5.1 Build the AgniOne unit.
Build script require the destination path to copy the binaries

Unit binaries will be copied to ~/AgniOneFM/AgniOne/apps/units folder
Units config file will be copied to ~/AgniOneFM/AgniOne/apps/config/demohttp folder

```
chmod 775 ./build.sh
./build.sh ~/AgniOneFM/agnione
```

### Run AgniOne

1. run the last build AgniOne Application framework + AgniOne PlugIns + AgniOne Units
```
cd ~/AgniOneFM/AgniOne
./run.sh
```

2. Monitor application log
```
tail -f ~/AgniOneFM/agnione/log/*
```

3. REST Monitor API <br>
   config/core.config contains the port for HTTP monitoring & Web Socket monitoring.
   ```
   "http_monitor": {
        "host": "0.0.0.0",
        "port": 8080,
        "enable": 1
      },
      "ws_monitor": {
        "host": "0.0.0.0",
        "port": 2345,
        "enable": 1
   ```
   All the HTTP REST endpoint will be hosted at http://localhost:8080
   
   Check liveness -> http://localhost:8080/live
   
    *** For detail monitoring please refer [Monitoring Guide](./README_Monitoring.md)
   

### Deploy Binaries

1. Run the last built AgniOne (Application framework + PlugIns + Units)
2. Use deploy script as below.
```
cd ~/AgniOneFM/AgniOne
./deploy.sh <TARGET-FOLDER>
```

## To DO
Please refer to To [TODO](./README.md#todo)

## Support
Please refer to [Support](./README.md#support) section.



