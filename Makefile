
exporter:
	go build -ldflags "-s -w" -o dist/exporter cmd/exporter/exporter.go


migrator:
	go build -ldflags "-s -w" -o dist/migrator cmd/migrator/migrator.go
