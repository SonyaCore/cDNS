APP_NAME=cdns
CMD_DIR=cmd/cdns

.PHONY: all build run clean

all: build

build:
	go build -o $(APP_NAME) $(CMD_DIR)/main.go

run: build
	./$(APP_NAME)

clean:
	rm -f $(APP_NAME)
