SHELL=/bin/sh

db:
	docker compose -f docker-compose.yml -f docker-compose.test.yml up

rebuild:
	npm install
	$(call compile)
	./node_modules/.bin/webpack --mode=development

build:
	npm ci
	$(call compile)
	./node_modules/.bin/webpack --mode=development

build_prod:
	npm ci
	$(call compile)
	./node_modules/.bin/webpack --mode=production

compile:
	cd ballot && go build -o ../bin/ballot

run:
	$(call compile)
	@cd ballot && ENV=development ../bin/ballot

test:
	$(call compile)
	@cd ballot && REDIS_URL=redis://localhost:6380 go test -v

logs:
	@docker-compose logs -f
