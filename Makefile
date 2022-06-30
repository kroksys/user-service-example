up:
	docker-compose up --remove-orphans -d

down:
	docker-compose down --remove-orphans

gen:
	buf mod update pkg/pb  
	buf generate pkg/pb

.Phony: up down gen