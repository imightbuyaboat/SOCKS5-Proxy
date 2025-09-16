include .env
export $(shell sed 's/=.*//' .env)

.PHONY: socks5-server-up relay-server-up socks5-server-down relay-server-down

socks5-server-up:
	docker run -d \
		--name db \
		--network host \
		-e POSTGRES_USER=$$POSTGRES_USER \
		-e POSTGRES_PASSWORD=$$POSTGRES_PASSWORD \
		-e POSTGRES_DB=$$POSTGRES_DB \
		-v $$HOME/docker/volumes/postgres:/var/lib/postgresql/data \
		postgres

	docker build -f socks5-server/Dockerfile \
		-t socks5-server .

	docker run -d \
		--network host \
		--name socks5-server-container socks5-server

relay-server-up:
	docker build -f relay-server/Dockerfile \
		-t relay-server .

	docker run -d \
		--network host \
		--name relay-server-container relay-server

socks5-server-down:
	docker stop socks5-server-container || true
	docker rm socks5-server-container || true

	docker stop db || true
	docker rm db || true

relay-server-down:
	docker stop relay-server-container || true
	docker rm relay-server-container || true