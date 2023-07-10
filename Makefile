BUILD_PATH := ./target
PREFIX := /usr/local/bin
BIN := golottie
RENDER_PATH = ./misc/test_render

.PHONY build install test clean covsh lint gobadge test_cmd:

$(BUILD_PATH)/$(BIN) build:
	mkdir -p $(BUILD_PATH)
	go build -o $(BUILD_PATH) ./...

$(PREFIX)/$(BIN) install: $(BUILD_PATH)/golottie
	install -m 557 $(BUILD_PATH)/$(BIN) $(PREFIX)/$(BIN)

clean:
	go clean
	rm -f $(BUILD_PATH)/*
	rm -rf $(RENDER_PATH)
	
coverage.out test:
	go test -cover -coverprofile coverage.out

covsh: coverage.out 
	env GOCOVSH_THEME=macchiato
	gocovsh 
	
lint:
	golangci-lint run
	
benchmark:	
	go test -bench=. -run=^# ./...

test_cmd: $(BUILD_PATH)/$(BIN)
	mkdir -p $(RENDER_PATH)
	./target/golottie -i misc/test.json -o $(RENDER_PATH)/%04d.png -w 600 -h 600 -c 4

test_cmd_gpu: $(BUILD_PATH)/$(BIN)-gpu
	mkdir -p $(RENDER_PATH)
	./target/golottie-gpu -i misc/test.json -o $(RENDER_PATH)/%04d.png -w 600 -h 600 -c 4

gobadge:
	go test -covermode=count -coverprofile=coverage.out
	go tool cover -func=coverage.out -o=coverage.out
	gobadge -filename=coverage.out