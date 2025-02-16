build:
	go build -o ./bin/block_chain

run: build
	./bin/block_chain

test:
	go test ./...
