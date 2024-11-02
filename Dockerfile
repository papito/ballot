#----------------------------
FROM node:18.20.4-alpine AS build_ui

COPY ./ballot-ui /app
WORKDIR /app
RUN npm ci
RUN npm run build

FROM golang:1.21 AS build_service
COPY . /app

WORKDIR /app/ballot
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o ballot

FROM alpine:latest AS runtime

RUN apk add --no-cache redis

#
WORKDIR /app
RUN mkdir /app/server
COPY --from=build_service /app/ballot/ballot /app/server/ballot
COPY --from=build_ui /app/dist/ ./ballot-ui/dist/
COPY entrypoint.sh ./

EXPOSE 8080
WORKDIR /app
CMD ["./entrypoint.sh"]
