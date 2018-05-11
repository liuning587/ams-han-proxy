proto_electricity = proto/electricity/electricity.pb.go

.PHONY: all client server clean
all: client server

client: $(proto_electricity)
	GOOS=linux GOARCH=arm GOARM=6 go build

server: $(proto_electricity)

clean:
	rm $(proto_electricity)

$(proto_electricity): %.pb.go : %.proto
	protoc $^ --go_out=plugins:.
