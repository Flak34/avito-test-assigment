.PHONY: run-integration-test
run-integration-test:
	go test -count=1 -tags integration ./test/...

.PHONY: start-test-env
start-test-env:
	cd ./test && docker compose up -d

.PHONY: truncate-test-data
truncate-test-data:
	psql -c "TRUNCATE banner CASCADE;" postgresql://test:test@localhost:5437/test

.PHONY: start-service
start-service:
	docker compose up -d