# we will put our integration testing in this path
INTEGRATION_TEST_PATH?=./it

# set of env variables that you need for testing
ENV_LOCAL_TEST=\
  POSTGRES_PASSWORD=password \
  POSTGRES_DB=db \
  POSTGRES_USER=user

# this command will start a docker components that we set in docker-compose.yml
docker.start.components:
	docker-compose  -f test-docker-compose.yml up --build -d

# shutting down docker components
docker.stop:
	docker-compose -f test-docker-compose.yml down

# this command will trigger integration test
# INTEGRATION_TEST_SUITE_PATH is used for run specific test in Golang, if it's not specified
# it will run all tests under ./it directory
test.integration:
	go test -tags=integration $(INTEGRATION_TEST_PATH) -count=1 -run=$(INTEGRATION_TEST_SUITE_PATH)

# this command will trigger integration test with verbose mode
test.integration.debug:
	go test -tags=integration $(INTEGRATION_TEST_PATH) -count=1 -v -run=$(INTEGRATION_TEST_SUITE_PATH)
