BINARY := weather
CMD    := ./cmd/weather

.PHONY: build run test install clean

build:
	go build -o $(BINARY) $(CMD)

run:
	go run $(CMD)

test:
	go test ./...

install:
	go install $(CMD)

clean:
	rm -f $(BINARY)
