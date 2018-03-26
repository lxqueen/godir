all:
	@# Build
	@mkdir -p ./out
	go get github.com/BurntSushi/toml
	go build -o "./out/godir" ./src/*.go
	@echo "Done"

clean:
	rm -rf ./out
