#----------------------------
FROM node:12.10.0-alpine as build_ui

COPY . /app
WORKDIR /app
RUN npm ci
RUN ./node_modules/.bin/webpack --mode=production


#----------------------------
FROM golang:1.12 AS build_service
COPY . /app

WORKDIR /app/ballot
RUN go get -d
RUN go build -o ballot

#----------------------------
FROM alpine:latest AS runtime
# the libc from the build stage is not the same as the alpine libc
# create a symlink to where it expects it since they are compatable. https://stackoverflow.com/a/35613430/3105368
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

WORKDIR /app
RUN mkdir /app/server
COPY --from=build_service /app/ballot/ballot /app/server/ballot
COPY --from=build_ui /app/ui/dist/ ./ui/dist/
COPY --from=build_ui /app/ui/templates/ ./ui/templates/
COPY start.sh ./

RUN apk add --no-cache redis

EXPOSE 8080

WORKDIR /app
CMD ["./start.sh"]
