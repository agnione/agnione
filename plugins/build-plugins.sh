#! /bin/bash
echo "./build_plugins.sh <BASE_PATH_TO_DEPLOY>"
echo "eg ./build_plugins.sh  /usr/src/app/zappfm/"
echo "deployment path -- > $1"
echo ""
echo "Building plugins......"
echo ""

echo "building http plugin....."
cd http/ahttpclient
./build.sh $1
echo "building http plugin..... DONE"

cd ..
cd ..

echo "building logger plugin....."
cd ./logger/alogger
./build.sh $1
echo "building logger plugin..... DONE"

cd ..
cd ..

echo "building mailer plugin....."
cd mailer/amailer
./build.sh $1
echo "building mailer plugin.....DONE"

cd ..
cd ..

echo "building reids MQ plugin....."
cd mq/redis/aredisclient
./build.sh $1
echo "building reids MQ plugin.....DONE"

cd ..

echo "building reids cluster MQ plugin....."
cd zredisclusterclient
./build.sh $1
echo "building reids MQ plugin.....DONE"

cd ..
cd ..
cd ..

echo "building websocket plugin....."
cd websocket/zwsclient
./build.sh $1
echo "building websocket plugin.....DONE"


echo "Building plugins......COMPLETED"


