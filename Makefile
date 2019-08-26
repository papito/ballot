SHELL=/bin/sh
CONTAINER_NAME=test_redis

define up_if_down
@if [ ! `docker ps -q -f name=$(CONTAINER_NAME)` ]; then echo "Bringing up containers\n--------" && docker-compose up -d; fi
endef

define compile
	cd ballot && go build -o ../bin/ballot
endef

compile:
	$(call compile)

run:
	$(call up_if_down)
	$(call compile)
	@cd ballot && ENV=development ../bin/ballot

up:
	$(call up_if_down)

down:
	@docker-compose down

test:
	$(call up_if_down)
	$(call compile)
	@cd ballot && REDIS_URL=redis://localhost:6380 go test -v

ps:
	@docker-compose ps

logs:
	@docker-compose logs -f

