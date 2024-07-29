TARGET=lazyssh
.PHONY: install
install:
	go build -v -o ${TARGET} main.go && cp ${TARGET} /Users/yejiaxin/Documents/code/GOPATH/bin