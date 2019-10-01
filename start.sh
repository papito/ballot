#!/bin/sh

cd server
redis-server --daemonize yes
./ballot


