SHELL=/bin/sh
CONTAINER_NAME=test_redis

define up_if_down
@if [ ! `docker ps -q -f name=$(CONTAINER_NAME)` ]; then echo "Bringing up containers\n--------" && docker-compose up -d; fi
endef

define compile
	@cd ballot && go build -o ../bin/ballot
endef

build:
	npm ci
	$(call compile)
	./node_modules/.bin/webpack --mode=development

build_prod:
	npm ci
	$(call compile)
	./node_modules/.bin/webpack --mode=production

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
	@cd ballot && REDIS_URL=localhost:6380 go test -v
	@echo "-------\nRun 'make down' to stop test containers..."

ps:
	@docker-compose ps

logs:
	@docker-compose logs -f

