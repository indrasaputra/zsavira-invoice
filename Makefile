.PHONY: tidy
tidy:
	GO111MODULE=on go mod tidy

.PHONY: compile
compile:
	GO111MODULE=on CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o invoice cmd/web/main.go

.PHONY: build
build:
	docker build --no-cache -t indrasaputra/zsavira-invoice:0.0.1 -f Dockerfile .

.PHONY: run
run:
	docker run -p 8000:8000 indrasaputra/zsavira-invoice:0.0.1