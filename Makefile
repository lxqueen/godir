.PHONY: clean test install

all:
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
	@mkdir -p $(HOME)/.config/godir/ || echo "Could not create configuration directory. Please create the directory $(HOME)/.config/godir/"
	@cp ./src/config.toml.example $(HOME)/.config/godir/config.toml || echo "Could not copy config.toml.example from ./src/ to $(HOME)/.config/godir! Please create or copy it there, or specify it on the command line when godir is run using the -c flag."
	@echo "Done"

clean:
	rm -rf ./out
	rm ./src/gdx.db
	@echo "Done"

test: all
	go test ./...
