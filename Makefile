all:
	env GOOS=linux GOARCH=amd64 go build -o json2rdf.linux-amd64
	env GOOS=windows GOARCH=amd64 go build -o json2rdf.windows-amd64
	env GOOS=darwin GOARCH=amd64 go build -o json2rdf.darwin-amd64
