.PHONY: clean test install

all:
	@# Build
	@mkdir -p ./out
	go get github.com/BurntSushi/toml
	go get github.com/kalafut/imohash/...
	go get github.com/otiai10/copy

	go get github.com/mattn/go-sqlite3
	@go install github.com/mattn/go-sqlite3

	@echo "go build -o \"./out/godir\""
	@go build -o "./out/godir" ./src/*.go || echo "Build Failed"
	@echo "Done"

install:
	@cp ./out/godir /usr/local/bin/godir || echo "Make install must be run as root."

clean:
	rm -rf ./out
	rm ./src/gdx.db
	@echo "Done"

test: all
	go test ./...
