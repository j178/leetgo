install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

gen:
	cd ./leetcode && protoc --go_out=. --go_opt=paths=source_relative -I . leetcode.proto
