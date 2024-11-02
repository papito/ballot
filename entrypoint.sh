#!/bin/sh

cd server || exit

redis-server --daemonize yes
redis-server

export REDIS_URL=redis://localhost:6379

./ballot
