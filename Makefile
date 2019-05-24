install:
	go install ./cmd/...

build-for-linux:
	GOOS=linux GOARCH=amd64 go build -o dist/telehello-linux-amd64 ./cmd/telehello

transfer:
	curl --upload-file dist/telehello-linux-amd64 https://transfer.sh/telehello
