test: linter standard

clean:
	rm -f serve/files.go

linter: clean
	# ./travis/test.sh linter

standard: generate
	./travis/test.sh standard

generate:
	go generate $$(go list ./... | grep -v /vendor/)
