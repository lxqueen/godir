all:
	@# Build
	@mkdir -p ./out
	go get github.com/BurntSushi/toml
	go get github.com/OneOfOne/xxhash
	go get gopkg.in/cheggaaa/pb.v1
	go build -o "./out/godir" ./src/*.go
	@echo "Done"

clean:
	rm -rf ./out

testrun:
	out/godir -v -o ./log.out -c ./src/config.toml.example ./test
