up:
	docker-compose up --remove-orphans -d

down:
	docker-compose down --remove-orphans

.Phony: up down