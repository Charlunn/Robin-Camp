.PHONY: docker-up docker-down test-e2e

ENV ?= dev
COMPOSE_FILE := docker-compose.$(ENV).yml

# 在解析阶段直接检查文件是否存在（跨 shell、跨平台）
ifeq ($(strip $(wildcard $(COMPOSE_FILE))),)
$(error Unknown environment '$(ENV)'. Expected file $(COMPOSE_FILE) to exist)
endif

ifeq ($(ENV),dev)
DOWN_FLAGS ?= -v
else
DOWN_FLAGS ?=
endif

docker-up:
	docker compose -f $(COMPOSE_FILE) up -d --build

docker-down:
	docker compose -f $(COMPOSE_FILE) down $(DOWN_FLAGS)

test-e2e:
	ENV_FILE=.env ./e2e-test.sh
