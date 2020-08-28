all:
	export GO111MODULE=off
	export GOPRIVATE=github.com
	go get -u -v
	github.com/donomii/nucular
	go build .
	go build -o shonkyTerm ./term
	go build -o shonkyEd3 ./v3
	go build -o shonkyEd2 ./v2
	go build -o textOnPic ./textpic/
