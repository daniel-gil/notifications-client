# Server

This test server launches a web server implemented using the go web framework `gin-gonic` and is listening to the route `/api/notifications`.

## Start
```bash
$ go run main.go 
2019/04/08 17:24:07 Server configuration: errorRatePercentage=0%
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] POST   /api/notifications        --> main.main.func1 (3 handlers)
[GIN-debug] Listening and serving HTTP on :9090
```

## Forcing errors
For testing purposes we can configure the server for introducing a certain error percentage using the `error` flag:

```bash
$ go run main.go -error=25
2019/04/08 17:28:22 Server configuration: errorRatePercentage=25%
[GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.

[GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
 - using env:	export GIN_MODE=release
 - using code:	gin.SetMode(gin.ReleaseMode)

[GIN-debug] POST   /api/notifications        --> main.main.func1 (3 handlers)
[GIN-debug] Listening and serving HTTP on :9090
```

## Handling requests
The server will display a log entry when handling new requests:
```bash
[GIN] 2019/04/08 - 17:29:48 | 200 |     144.664µs |             ::1 | POST     /api/notifications
```