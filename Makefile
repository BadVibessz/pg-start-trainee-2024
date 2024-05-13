# testing vars
export TEST_CONTAINER_NAME=test_db
export TEST_DBSTRING=postgresql://postgres:postgres@localhost:5433/test?sslmode=disable
export TEST_GOOSE_DRIVER=postgres
export TEST_DOCKER_PORT=5433

export DOCKER_IMAGE_NAME=pg_start_test_trainee_image

run:
	docker compose up

test.integration:
	docker run --rm -d -p $$TEST_DOCKER_PORT:5432 --name $$TEST_CONTAINER_NAME -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=test postgres

	sleep 2 # wait for postgres to run in docker container, todo: bad practice, use go-migrate instead?

	# [command] || true prevents the script to stop even if error occurred executing command, so newly created docker container will be deleted anyway

	goose -dir ./db/migrations $$TEST_GOOSE_DRIVER $$TEST_DBSTRING up || true # apply migrations
	go test -v ./tests/* || true # run tests

	docker stop $$TEST_CONTAINER_NAME # stop docker container and then delete it

up_test_db:
	docker run --rm -d -p $$TEST_DOCKER_PORT:5432 --name $$TEST_CONTAINER_NAME -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=test postgres

	sleep 2

	goose -dir ./db/migrations $$TEST_GOOSE_DRIVER $$TEST_DBSTRING up || true