![mxt](https://github.com/txn2/mxt/blob/master/mast.jpg?raw=true)
[![query Release](https://img.shields.io/github/release/txn2/mxt.svg)](https://github.com/txn2/mxt/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/txn2/mxt)](https://goreportcard.com/report/github.com/txn2/mxt)
[![GoDoc](https://godoc.org/github.com/txn2/mxt?status.svg)](https://godoc.org/github.com/txn2/mxt)
[![Docker Container Image Size](https://shields.beevelop.com/docker/image/image-size/txn2/mxt/latest.svg)](https://hub.docker.com/r/txn2/mxt/)
[![Docker Container Layers](https://shields.beevelop.com/docker/image/layers/txn2/mxt/latest.svg)](https://hub.docker.com/r/txn2/mxt/)

Endpoint data transformation shim using [Tengo](https://github.com/d5/tengo) for scripting.


## Demo

Start fake API server:
```bash
docker-compose up
```

Run mxt:
```bash
go run ./cmd/mxt.go -config=./cfg/simple.yml
```

Get endpoint:
```bash
curl http://localhost:8080/get/twentyfour
```


## Development

### Test Release

```bash
goreleaser --skip-publish --rm-dist --skip-validate
```

### Release

```bash
GITHUB_TOKEN=$GITHUB_TOKEN goreleaser --rm-dist
```