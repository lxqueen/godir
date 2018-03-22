all:
	@# Build
	@mkdir -p ./out
	go build -o "./out/godir" ./src/*.go
	@echo "Done"

clean:
	rm -rf ./out
