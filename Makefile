SHELL=/bin/sh

db:
	docker compose -f docker-compose.yml -f docker-compose.test.yml up


compile:
	cd ballot && go build -o ../bin/ballot

start:
	$(call compile)
	@cd ballot && ENV=development ../bin/ballot

test:
	$(call compile)
	@cd ballot && REDIS_URL=redis://localhost:6380 go test -v

logs:
	@docker-compose logs -f
