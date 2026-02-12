.PHONY: genkeys run migrate test

genkeys:
	rm -rf keys && \
	mkdir keys && \
	openssl genpkey -algorithm ED25519 -outform pem -out keys/auth.ed && \
	openssl pkey -in keys/auth.ed -pubout > keys/auth.ed.pub

run:
	go run cmd/main.go keys/auth.ed keys/auth.ed.pub

migrate:
	migrate -source file://./migrations -database postgresql://test:test@localhost:5432/test?sslmode=disable up

test:
	APIURL=http://localhost:8080/api ./api/run-api-tests.sh
