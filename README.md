# x-gocommon
Shared Go library for HTTP, gRPC, and worker processes.
`context.Context` is the in-process source of truth. HTTP headers and gRPC metadata are only used at process boundaries.
```bash
go get github.com/huponx/x-gocommon@latest
```
Requires Go 1.26+.
## Packages
| Package | Role |
|---|---|
| [`requestctx`](#requestctx) | Identity on context; header/metadata mapping |
| [`logging`](#logging) | Zap; `FromCtx` adds request fields |
| [`config`](#config) | Env load + reusable blocks (`Server`, `Log`, `Postgres`, `Redis`, `GRPCClient`) |
| [`grpcserver`](#grpcserver) | Server + interceptors, gRPC health, graceful stop |
| [`grpcclient`](#grpcclient) | Dial + interceptors (ctx → metadata, unary timeout, logging) |
| [`httpx`](#httpx) | Gin middleware: logger, request context, access log |
| [`healthz`](#healthz) | Gin `/healthz` + `/readyz` |
| [`worker`](#worker) | Signals, in-flight drain, optional stdlib probes |
| [`apperror`](#apperror) | Typed errors → gRPC `status` + HTTP JSON |
## `requestctx`
```go
type Values struct {
	CorrelationID, UserID, TenantID, AgencyID string
	RequestID, TraceID, SpanID                string
}
```
| Boundary | API |
|---|---|
| In-process | `WithValues`, `From`, `SetUserID`, `EnsureIDs` |
| HTTP in | `FromHTTP` — reads `X-Correlation-ID`, `X-Request-ID`, `X-Tenant-ID`, `X-Agency-ID`, `X-Trace-ID`, `X-Span-ID`. **Does not read user id.** Resolve auth, then `SetUserID`. |
| HTTP out | `ToHTTPHeaders` |
| gRPC in | `HydrateIncoming` / `FromMD` (`correlation-id`, `user-id`, …) |
| gRPC out | `AppendOutgoing` — copies `Values` onto outgoing metadata without overwriting keys already set |
`EnsureIDs` generates `correlation_id` and `request_id` when missing.
## `logging`
```go
log, err := logging.New(logging.Config{
	Level:    "info",   // zap level
	Encoding: "console", // or "json"
	Service:  "http-server",
	Env:      "dev",
})
defer logging.Sync(log)
ctx = logging.WithCtx(ctx, log)
logging.FromCtx(ctx).Info("ok") // adds correlation_id, user_id, request_id, tenant_id when present
```
## `config`
Embed blocks on your service config and load from env:
```go
type Config struct {
	Server config.Server
	Log    config.Log
	GRPC   config.GRPCClient
}
cfg := config.MustLoad[Config]()
```
| Block | Env (defaults) |
|---|---|
| `Server` | `HOST` (`0.0.0.0`), `PORT` (`8080`), `SHUTDOWN_TIMEOUT` (`15s`) — `Addr()` |
| `Log` | `LOG_LEVEL` (`info`), `LOG_ENCODING` (`console`), `SERVICE_NAME` (`app`), `ENV` (`dev`) |
| `Postgres` | `POSTGRES_HOST/PORT/USER/PASSWORD/DB/SSLMODE` — `URL()` |
| `Redis` | `REDIS_HOST/PORT/PASSWORD/DB` — `URL()` |
| `GRPCClient` | `GRPC_TARGET` (`localhost:9090`), `GRPC_TIMEOUT` (`5s`), `GRPC_INSECURE` (`true`) |
## `httpx`
Gin middleware. Typical order:
```go
r.Use(gin.Recovery())
r.Use(httpx.Logger(log))
r.Use(httpx.RequestContext())
// app auth → requestctx.SetUserID
r.Use(httpx.AccessLog())
```
## `grpcserver`
Interceptors (unary + stream): recovery → hydrate `requestctx` → optional auth hook → logging.
```go
srv := grpcserver.New(
	grpcserver.WithLogger(log),
	grpcserver.WithReflection(dev),
	grpcserver.WithAuth(func(ctx context.Context) (context.Context, error) {
		return requestctx.SetUserID(ctx, userID), nil
	}),
)
pb.RegisterFooServer(srv.GRPC, impl)
srv.SetServing(pb.Foo_ServiceDesc.ServiceName, true)
// on signal:
srv.Stop(shutdownCtx) // health Shutdown, then GracefulStop; hard Stop if ctx expires
```
Options: `WithHealth`, `WithMaxMsgSize`, `WithServerOptions`.
## `grpcclient`
```go
conn, err := grpcclient.Dial(ctx, cfg.GRPC.Target,
	grpcclient.WithLogger(log),
	grpcclient.WithTimeout(cfg.GRPC.Timeout),
	grpcclient.WithInsecure(cfg.GRPC.Insecure),
)
```
Unary interceptors copy `requestctx` onto outgoing metadata, apply a per-call timeout, and log. TLS via `WithTLS` / `WithInsecure(false)`.
## `healthz`
Gin handlers. JSON: `{"status":"ok","checks":{"name":"ok"}}`. Failed checks → `503` + `"unavailable"`.
```go
r.GET("/healthz", healthz.Live())
r.GET("/readyz", healthz.Ready(healthz.GRPC("grpc", conn, "")))
```
`Check` is `Name` + `Fn(ctx) error`. Reuse the same `Check` values with `worker.NewHealth`.
## `worker`
No queue client. Service owns consume/ack; this package owns process lifecycle.
1. **`Signals()`** — cancel on SIGINT/SIGTERM. Stop accepting new jobs; in-flight work is not cancelled yet.
2. **`Runner`** — `Go(ctx, fn)` runs a job. Job context keeps parent values (`requestctx`, …) and is cancelled only when **`Drain`'s deadline expires**. After drain starts, `Go` returns `false` (nack/requeue).
3. **`Health`** — optional stdlib `/healthz` + `/readyz`. `SetReady(false)` before drain so readiness fails first.
```go
r := worker.New(worker.WithLogger(log))
h := worker.NewHealth(cfg.Server.Addr(), healthz.Check{Name: "redis", Fn: pingRedis})
go func() { _ = h.ListenAndServe() }()
ctx, stop := worker.Signals()
defer stop()
go consume(ctx, r) // receive → r.Go(jobCtx, handle)
<-ctx.Done()
h.SetReady(false)
drainCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
defer cancel()
_ = r.Drain(drainCtx)
_ = h.Shutdown(drainCtx)
```
## `apperror`
gRPC `codes.Code` plus a message. Implements `GRPCStatus()` so handlers can `return err` to gRPC.
```go
return apperror.NotFound("user not found")
apperror.WriteHTTP(c, err) // {"error":{"code":"NotFound","message":"user not found"}}
```
Helpers: `InvalidArgument`, `NotFound`, `Unauthenticated`, `PermissionDenied`, `AlreadyExists`, `FailedPrecondition`, `Unavailable`, `Internal`, `Wrap`, `From`.
HTTP mapping includes `InvalidArgument` → 400, `Unauthenticated` → 401, `NotFound` → 404, `Unavailable` → 502, `DeadlineExceeded` → 504, unknown → 500.
## Identity flow
HTTP edge hydrates context from headers, the app sets `user_id` after auth, the gRPC client copies values into metadata, the gRPC server hydrates context again. The same IDs show up on access logs and `logging.FromCtx`.
Workers should copy `Values` from the enqueue payload into `Go`'s `ctx` if they need the same log fields.
## Local development
This module is used from sibling services via `replace` / `go.work`. After publishing:
```bash
go get github.com/huponx/x-gocommon@v0.1.0
```
```bash
go test ./...