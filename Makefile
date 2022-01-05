compose-up:
	docker compose up -d --build --remove-orphans
compose-down:
	docker compose down --remove-orphans
