#! /bin/bash

#########################################################################################
#   Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 15/01/2025
#   Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
#						Licensed under the Apache License, Version 2.0 (the "License");
#						you may not use this file except in compliance with the License.
#						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
#########################################################################################

VESRION=1.0.0
SOURCE=./app.go
BINARY=./agnione.app

BuildTime=`date`
BuildGoVersion=`go version`

LDFLAGS=" -s -w -X 'agnione.appfm/src/build.Version=${VESRION}' \
-X 'agnione.appfm/src/build.User=$(id -u -n)' \
-X 'agnione.appfm/src/build.Time=${BuildTime}' \
-X 'agnione.appfm/src/build.BuildGoVersion=${BuildGoVersion}' "


cd ./src
echo "clean old binaries....."
rm ${BINARY}
echo "clean old binaries ......... DONE"

echo "building AgniOne Application Framework ........"

go build -v -ldflags="${LDFLAGS}" -o ${BINARY} ${SOURCE}

echo "building AgniOne Application Framework ............. DONE"


