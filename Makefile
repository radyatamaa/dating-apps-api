swagger_documentation:
	swag init -g ./main.go --output swagger

run:
	docker compose -f "docker-compose.yml" up -d --build

stop:
	docker compose -f "docker-compose.yml" down

test: 
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out