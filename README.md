![mxt](./mast.jpg)

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

