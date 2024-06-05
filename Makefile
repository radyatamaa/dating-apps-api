swagger_documentation:
	swag init -g ./main.go --output swagger

run:
	docker compose -f "docker-compose.yml" up -d --build

stop:
	docker compose -f "docker-compose.yml" down
