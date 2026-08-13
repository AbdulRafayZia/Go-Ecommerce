# Documentation Updates Summary

## Overview
Updated all project documentation to reflect the change from gRPC to REST/HTTP for service-to-service communication.

## Files Updated

### 1. ADR 0001 - Communication Protocol
- **Old**: `0001-why-grpc-over-rest-internally.md`
- **New**: `0001-use-rest-for-service-communication.md`
- **Changes**: Complete rewrite explaining REST/HTTP + OpenAPI choice

### 2. ADR 0003 - Hexagonal Architecture
- Updated code examples from gRPC to HTTP
- Changed `grpc/` adapter folder to `http/`
- Updated Proto → JSON mapping references
- Changed `grpcutil` to `httputil`

### 3. README.md
- Updated "Overview" section to mention REST/HTTP + OpenAPI
- Changed tech stack from "gRPC + Protocol Buffers" to "REST/HTTP + OpenAPI"
- Updated development tools from "Buf" to "oapi-codegen + Chi Router"
- Changed "make proto" to "make api-gen"
- Updated project structure diagrams
- Updated hexagonal architecture example

### 4. Architecture Documentation (docs/architecture.md)
- Updated communication patterns section
- Changed synchronous communication from gRPC to REST/HTTP
- Updated "Why REST/HTTP?" rationale
- Changed references from "gRPC call" to "HTTP call"
- Updated observability section

### 5. Makefile
- Removed `proto`, `proto-lint`, `proto-breaking` commands
- Added `api-gen` command for OpenAPI code generation
- Added `api-lint` command for OpenAPI spec validation

## Important Note: gRPC Still Used for Tracing

**The `pkg/tracing/tracing.go` file still uses gRPC** - this is intentional and correct!

### Why?
gRPC is used in the tracing package to send telemetry data to the OTLP collector (Jaeger). This is:
- **NOT** service-to-service communication
- **IS** the standard protocol for OpenTelemetry trace export
- Required by the OpenTelemetry specification

### Clarification Added
Added comments in `tracing.go`:
```go
// Note: gRPC is used here ONLY for OTLP export to Jaeger/collector
// This is NOT the same as using gRPC for service-to-service communication
```

## Architecture Decision Summary

| Aspect | Before (gRPC) | After (REST) |
|--------|---------------|--------------|
| **Protocol** | gRPC + Protocol Buffers | REST/HTTP + JSON |
| **Contract** | .proto files | OpenAPI 3.0 specs |
| **Code Gen** | protoc + buf | oapi-codegen |
| **Adapter** | internal/adapters/grpc/ | internal/adapters/http/ |
| **Router** | grpc.Server | chi.Mux |
| **Errors** | grpc.Status codes | HTTP status codes |
| **Middleware** | gRPC interceptors | HTTP middleware |

## Benefits of This Change

1. **Simplicity**: REST is more widely understood
2. **Debugging**: JSON is human-readable vs binary protobuf
3. **Tooling**: curl, Postman, browser DevTools work out of the box
4. **Portfolio**: More relatable to interviewers (most companies use REST)
5. **Type Safety**: Still maintained via oapi-codegen

## Future Migration Path

If gRPC is needed later for specific high-throughput services:
1. Hexagonal architecture makes this easy
2. Only replace HTTP adapter with gRPC adapter
3. Domain and application layers remain unchanged
4. Can run both protocols in parallel

This demonstrates **architectural flexibility** - a key senior engineer skill!
