.PHONY: clean test testrun

all:
	@# Build
	@mkdir -p ./out
	go get github.com/BurntSushi/toml
	go get github.com/kalafut/imohash/...
	go get github.com/otiai10/copy

	go get github.com/mattn/go-sqlite3
	@go install github.com/mattn/go-sqlite3

	go build -o "./out/godir" ./src/*.go
	@echo "Done"

clean:
	rm -rf ./out
	rm ./src/dir.gdx

testrun:
	out/godir -v -c ./src/config.toml.example ./test

test:
	go test ./...
