windows:
	go build -ldflags "-s -w" -o dist/exporter.exe cmd/exporter/exporter.go
	go build -ldflags "-s -w" -o dist/migrator.exe cmd/migrator/migrator.go


linux:
	go build -ldflags "-s -w" -o dist/exporter cmd/exporter/exporter.go
	go build -ldflags "-s -w" -o dist/migrator cmd/migrator/migrator.go
