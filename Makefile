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
	./target/golottie -i misc/test.json -o misc/render2/%04d.png -w 600 -h 600 -c 4

gobadge:
	go test -covermode=count -coverprofile=coverage.out
	go tool cover -func=coverage.out -o=coverage.out
	gobadge -filename=coverage.out