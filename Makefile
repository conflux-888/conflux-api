.PHONY: dev d admin-dev admin-install admin-build build run swg clean

# Full-stack dev: air (backend) + vite (frontend) concurrently
d:
	@trap 'kill 0' INT TERM EXIT; \
	 CORS_ALLOW_LOCALHOST=true air & \
	 (cd admin && pnpm run dev) & \
	 wait

dev:
	CORS_ALLOW_LOCALHOST=true air

admin-dev:
	cd admin && pnpm run dev

admin-install:
	cd admin && pnpm install

admin-build:
	cd admin && pnpm install --frozen-lockfile && pnpm run build

# Builds admin UI first so go:embed picks up the latest dist
build: admin-build
	go build -o ./tmp/main .

run:
	go run .

swg:
	swag init --parseInternal --parseDependency -o ./swagger --packageName swagger

clean:
	rm -rf tmp admin/dist admin/node_modules
