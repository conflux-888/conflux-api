dev:
	air

build:
	go build -o ./tmp/main .

run:
	go run .

swagger:
	swag init --parseInternal --parseDependency -o ./swagger --packageName swagger
