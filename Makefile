# we will put our integration testing in this path
INTEGRATION_TEST_PATH?=./it

all: docker.start test.integration docker.stop

# this command will start a docker components that we set in test-docker-compose.yml
docker.start:
	docker-compose -p test-seamless -f test-docker-compose.yml up --build -d --remove-orphans

# this command will trigger integration test with verbose mode
test.integration:
	go test -tags=integration $(INTEGRATION_TEST_PATH) -count=1 -v -run=$(INTEGRATION_TEST_SUITE_PATH)

# shutting down docker components
docker.stop:
	docker-compose -p test-seamless -f test-docker-compose.yml down
