# ADR 0001: Use REST/HTTP with OpenAPI for Service Communication

**Status**: Accepted

**Date**: 2024-01-15 (Updated: 2024-08-13)

**Deciders**: Architecture Team

## Context

We need to choose a communication protocol for service-to-service communication within our microservices architecture. The primary candidates are:

1. REST/HTTP with JSON
2. gRPC with Protocol Buffers
3. GraphQL

## Decision

We will use **REST/HTTP with JSON** for all service-to-service communication, using **OpenAPI 3.0** specifications as contracts and **oapi-codegen** for type-safe code generation.

## Rationale

### Advantages of REST/HTTP

**Simplicity & Familiarity**
- Well-understood by all developers
- Standard HTTP methods (GET, POST, PUT, DELETE)
- JSON is human-readable and easy to debug
- Universal tooling support (curl, Postman, browser DevTools)

**Debugging & Observability**
- Easy to inspect payloads (plain JSON)
- Can test endpoints with simple curl commands
- Logs are human-readable
- No special tools needed for debugging

**Strong Typing with OpenAPI**
- OpenAPI 3.0 specs serve as contracts between services
- oapi-codegen generates type-safe Go code from specs
- Compile-time type checking for request/response models
- Auto-generated server interfaces ensure implementation compliance

**Ecosystem & Tooling**
- Mature HTTP client libraries (standard library, resty)
- Extensive middleware ecosystem (chi, gorilla)
- OpenAPI spec can generate documentation automatically
- Easy integration with API gateways

**HTTP/2 Benefits**
- Modern HTTP/2 support in Go's net/http
- Multiplexing, header compression
- Server push capabilities
- Most of gRPC's performance benefits without the complexity

**Flexibility**
- Easy to version (v1, v2 paths)
- Incremental adoption of new endpoints
- Can add GraphQL layer later if needed
- Works natively in browsers

### Why Not gRPC?

While gRPC offers better performance in theory:

- **Complexity**: Requires protobuf knowledge, code generation setup
- **Debugging**: Binary format requires special tools (grpcurl, Bloom RPC)
- **Overkill**: For a portfolio/learning project, REST is more demonstrable
- **Browser Compatibility**: Requires gRPC-Web proxy
- **Learning Focus**: REST patterns are more transferable to most backend roles

### Why Not GraphQL?

- Adds complexity for simple CRUD operations
- N+1 query problem in distributed systems
- Better suited for frontend-to-backend (potential future addition)

## Implementation Strategy

### 1. OpenAPI Specifications

Each service maintains an `api/openapi.yaml`:

```yaml
openapi: 3.0.0
info:
  title: Product Service API
  version: 1.0.0
paths:
  /v1/products:
    get:
      summary: List products
      operationId: listProducts
      # ...
```

### 2. Code Generation

Use `oapi-codegen` to generate:
- Request/response types
- Server interface
- Chi route registration helpers

```bash
oapi-codegen -generate types,chi-server -package api api/openapi.yaml > internal/adapters/http/generated.go
```

### 3. HTTP Routing

Use **chi** for routing:
- Lightweight and idiomatic
- Middleware composition
- Request-scoped context
- Compatible with stdlib

### 4. Middleware Stack

Standard middleware for all services:
- Logging (request/response)
- Correlation ID
- OpenTelemetry tracing
- Recovery (panic handling)
- Request timeout
- CORS (for gateway)

### 5. Error Handling

Standardized error responses:

```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product with ID xyz not found",
    "details": {
      "product_id": "xyz"
    }
  }
}
```

## Consequences

### Positive

✅ **Simplicity**: Easier for juniors and new team members
✅ **Debuggability**: JSON is human-readable, easy to troubleshoot
✅ **Tooling**: Universal support (curl, Postman, browser)
✅ **Documentation**: OpenAPI generates interactive docs (Swagger UI)
✅ **Type Safety**: oapi-codegen provides compile-time checks
✅ **Flexibility**: Easy to add endpoints, version APIs
✅ **Portfolio Value**: More relatable to interviewers (most companies use REST)

### Negative

⚠️ **Performance**: JSON parsing slower than protobuf
- *Acceptable*: Rarely a bottleneck in typical workloads
- *Mitigation*: Use HTTP/2, connection pooling, caching

⚠️ **Payload Size**: JSON is more verbose than binary formats
- *Acceptable*: Bandwidth is cheap, readability is valuable
- *Mitigation*: GZIP compression, pagination

⚠️ **No Streaming**: REST doesn't support bi-directional streaming
- *Acceptable*: Can use Server-Sent Events (SSE) or WebSockets if needed
- *Not needed*: Current requirements don't require streaming

### Neutral

⚖️ **Code Generation**: Still using code generation (oapi-codegen vs protoc)
- Both approaches ensure type safety
- OpenAPI is more readable than .proto files

## Versioning Strategy

Use URL-based versioning:

```
/v1/products
/v1/products/{id}
/v2/products  (when breaking changes needed)
```

Benefits:
- Clear and explicit
- Easy to maintain multiple versions
- Clients can migrate at their own pace

## Migration Path (If Needed)

If we later need gRPC for specific high-throughput services:

1. Hexagonal architecture makes this easy
2. Only replace the HTTP adapter with gRPC adapter
3. Domain and application layers remain unchanged
4. Can run both REST and gRPC in parallel

This demonstrates **architectural flexibility** in interviews.

## Implementation Checklist

- [x] Remove proto tooling from foundation
- [ ] Create OpenAPI specs per service
- [ ] Set up oapi-codegen in Makefile
- [ ] Create HTTP middleware (logging, tracing, correlation)
- [ ] Implement standard error handling
- [ ] Document API patterns
- [ ] Create example service (Product Service)

## Related Decisions

- **ADR 0002**: Outbox Pattern for Event Publishing (unchanged)
- **ADR 0003**: Hexagonal Architecture for Service Structure (unchanged)

## References

- [OpenAPI Specification](https://swagger.io/specification/)
- [oapi-codegen](https://github.com/deepmap/oapi-codegen)
- [Chi Router](https://github.com/go-chi/chi)
- [RESTful API Design Best Practices](https://restfulapi.net/)
- [Microsoft REST API Guidelines](https://github.com/microsoft/api-guidelines)
