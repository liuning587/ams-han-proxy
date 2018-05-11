client_bin = ams-han-proxy-client
server_bin = ams-han-proxy-server
base_pkg = svenschwermer.de/ams-han-proxy
proto_electricity = proto/electricity/electricity.pb.go

.PHONY: all client server clean
all: client server

client: $(proto_electricity)
	GOOS=linux GOARCH=arm GOARM=6 go build -o $(client_bin) $(base_pkg)/client

server: $(proto_electricity)
	GOOS=linux GOARCH=amd64 go build -o $(server_bin) $(base_pkg)/server

clean:
	rm $(proto_electricity)

$(proto_electricity): %.pb.go : %.proto
	protoc $^ --go_out=plugins:.
