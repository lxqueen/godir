all:
	@# Build
	@mkdir -p ./out
	go get github.com/BurntSushi/toml
	go get github.com/OneOfOne/xxhash
	go get -u github.com/schollz/progressbar
	go build -o "./out/godir" ./src/*.go
	@echo "Done"

clean:
	rm -rf ./out

testrun:
	out/godir -v -o ./log.out -c ./src/config.toml.example ./test
