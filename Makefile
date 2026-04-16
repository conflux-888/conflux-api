.PHONY: dev build run swg

dev:
	air

build:
	go build -o ./tmp/main .

run:
	go run .

swg:
	swag init --parseInternal --parseDependency -o ./swagger --packageName swagger
