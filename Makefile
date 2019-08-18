.ONESHELL:

SHELL=/bin/sh
CONTAINER_NAME=test_redis

define up_if_down
	@if [ ! `docker ps -q -f name=$(CONTAINER_NAME)` ]; then docker-compose up -d; else echo "---"; fi
endef

define compile
	cd ballot/
	go build -o ../bin/ballot
endef

compile:
	$(call compile)

run:
	$(call up_if_down)
	$(call compile)
	ENV=development ../bin/ballot

up:
	$(call up_if_down)

down:
	@docker-compose down

test:
	$(call up_if_down)
	$(call compile)
	REDIS_URL=redis://localhost:6380 go test -v

status:
	@docker-compose ps

logs:
	@docker-compose logs -f

