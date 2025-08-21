#################################################################################################################
# Author        :   D. Ajith Nilantha de Silva ajithdesilva@gmail.com, contact@agnione.net
# Copyright     :   AgniOne.Net 2025
# Date Written  :   20/06/2025
# Class/module  :   AgniOne Framework Dockerization
# Objective     :   Use the Alpine amd64 latest as the production docker image for AgniOne Framework.
#                   Build the AgniOne Framework and create a docker image with core configuration files
#                   All the Unit + configuration will be mounted via host to the running container
#################################################################################################################
##  V2
## 0. docker image rm agnione/appfm:0.0.0.2
##
## 1. build alpine docker image
##      docker build --no-cache -t agnione/appfm:0.0.0.2 .
##
## 2 test image
##      docker run --name agnione_test -i -t agnione/appfm:0.0.0.2
##
## 3 remove test container
##      docker rm agnione_test
###################################################################################################################

## builder
FROM golang:1.24.4-alpine AS builder_fm

RUN apk --no-cache add build-base bash git

RUN mkdir -p /home/agnione/logs /home/agnione/apps /home/agnione/plugins /home/src

## clone & setup packages
WORKDIR /home/src

RUN git clone -b v2 https://github.com/agnione/libs.git
RUN mkdir  /usr/local/go/src/agnione
RUN mv  ./libs/v2 /usr/local/go/src/agnione/

## clone & build plugins
RUN git clone -b v2 https://github.com/agnione/plugins.git
WORKDIR /home/src/plugins
RUN chmod 775 ./build-plugins.sh
RUN ./build-plugins.sh /home/agnione

## build AgniOne framework
WORKDIR /home/src
RUN git clone -b v2 https://github.com/agnione/agnione.git
WORKDIR  /home/src/agnione
RUN chmod 775 ./build.sh
RUN ./build.sh

### build the AgniOne Docker Image
#### use official alpine image
FROM amd64/alpine:latest AS rumtime
LABEL version="0.0.0.2"
LABEL maintainer="ajithdesilva@gmail.com,contact@agnione.net"

## make folders for framework execution
## it is possible to mount host paths to these folders
RUN mkdir -p /home/agnione/logs /home/agnione/apps /home/agnione/certs

## copy the built AgniOne framework executable
COPY --from=builder_fm /home/src/agnione/src/agnione.app /home/agnione
RUN chmod 744 /home/agnione/agnione.app

## copy built plugins
COPY --from=builder_fm /home/agnione/plugins /home/agnione/plugins

# Add the framework configs
ADD ./config /home/agnione/config

ADD ./agnione.sh /home/agnione
RUN chmod 744 /home/agnione/agnione.sh /home/agnione/agnione.app

EXPOSE 8080-8081
WORKDIR /home/agnione

ENTRYPOINT  ["./agnione.sh"]
