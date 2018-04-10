all:
	@# Build
	@mkdir -p ./out
	go get github.com/BurntSushi/toml
	go get github.com/OneOfOne/xxhash
	go get github.com/otiai10/copy
	go build -o "./out/godir" ./src/*.go
	@echo "Done"

clean:
	rm -rf ./out

testrun:
	out/godir -v -c ./src/config.toml.example ./test
