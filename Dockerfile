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
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -mod=vendor -o ballot

#----------------------------
FROM alpine:latest AS runtime

RUN apk add --no-cache redis

WORKDIR /app
RUN mkdir /app/server
COPY --from=build_service /app/ballot/ballot /app/server/ballot
COPY --from=build_ui /app/ui/dist/ ./ui/dist/
COPY --from=build_ui /app/ui/templates/ ./ui/templates/
COPY start.sh ./

EXPOSE 8080
WORKDIR /app
CMD ["./start.sh"]
