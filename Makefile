coverage:
	go test -cover -coverprofile coverage.out

covsh: coverage
	GOCOVSH_THEME=macchiato gocovsh 

benchmark:	
	go test -bench=. -run=^# ./...
	
lint:
	golangci-lint run
	
build:
	go build -o target ./...

test_cmd: build
	rm -r misc/render
	mkdir misc/render
	./target/golottie -i misc/test.json -o misc/render/%04d.png -w 600 -h 600 -c 4