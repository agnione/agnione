#! /bin/bash

#########################################################################################
#   Author        :   D. Ajith Nilantha de Silva contact@agnione.net | 26/01/2024
#   Copyright     :   © 2024 D. Ajith Nilantha de Silva contact@agnione.net
#						Licensed under the Apache License, Version 2.0 (the "License");
#						you may not use this file except in compliance with the License.
#						You may obtain a copy of the License at http://www.apache.org/licenses/LICENSE-2.0
#########################################################################################

echo "Deploying application framework"
echo ""
echo "Creating target folders ....... "

mkdir -p $/1
mkdir -p $1/logs


echo "creating target folders ......... DONE "
echo ""

echo "copying binaries to folders ....... "
cp -r plugins $1/
cp -r config $1/
cp src/agni.app $1/
chmod -R 775 $1/agni.app

echo "copying Agni binaries to folders ....... DONE "

echo "copying Agni Units ......."
cp ./apps $1
echo "copying Agni Units ....... DONE "

echo ""
mkdir -p /var/log/agni
touch /var/log/agni/init.txt

echo "Deploying application framework ........ DONE"
echo ""

echo "To Run, please execute below command "
echo "$1/agni.app --main_path $1/  --app_path $1/apps/config/demohttp/app.config --log_path  /var/log/agni/ --cpu_count 5 --rest_port 8080 --ws_port 2345"



