test: linter test-standard test-gopherjs

clean:
	rm -f serve/files.go

linter: clean
	# ./travis/test.sh linter

test-standard: generate
	./travis/test.sh standard

test-gopherjs: generate
	rm -rf ${GOPATH}/pkg/*_js
	./travis/test.sh gopherjs

generate:
	go generate $$(go list ./... | grep -v /vendor/)
