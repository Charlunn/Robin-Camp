.PHONY: docker-up docker-down test-e2e

ENV ?= dev
COMPOSE_FILE := docker-compose.$(ENV).yml

ifeq ($(ENV),dev)
DOWN_FLAGS ?= -v
else
DOWN_FLAGS ?=
endif

docker-up:
	@if [ ! -f $(COMPOSE_FILE) ]; then echo "Unknown environment '$(ENV)'. Expected a docker-compose.$(ENV).yml file." >&2; exit 1; fi
	docker compose -f $(COMPOSE_FILE) up -d --build

docker-down:
	@if [ ! -f $(COMPOSE_FILE) ]; then echo "Unknown environment '$(ENV)'. Expected a docker-compose.$(ENV).yml file." >&2; exit 1; fi
	docker compose -f $(COMPOSE_FILE) down $(DOWN_FLAGS)

test-e2e:
	ENV_FILE=.env ./e2e-test.sh
