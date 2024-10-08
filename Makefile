app.docker.test.stop:
	docker compose --file ./tests/config/docker-compose.yml  down --remove-orphans

app.docker.test.start:
	docker compose --file ./tests/config/docker-compose.yml  up -d  

test: app.docker.test.start
		go test -count=1 -tags integration  ./tests/integration/

test.verbose: app.docker.test.start
		go test -count=1 -tags integration -v  ./tests/integration/

run.dev.watch: app.docker.test.start 
		air -c .air.toml

