#! /bin/bash

### rung with customer ports

./src/agnione.app --main_path ${PWD}/  --app_path ${PWD}/apps/config/demohttp/app.config --log_path ${PWD}/log/ --cpu_count 5 --rest_port 8080 --ws_port 2345