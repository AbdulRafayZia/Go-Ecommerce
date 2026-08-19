module gocommerce/services/order

go 1.25.0

require (
	github.com/go-chi/chi/v5 v5.3.1
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.10.9
	github.com/oapi-codegen/runtime v1.6.0
	github.com/segmentio/kafka-go v0.4.47
	gocommerce v0.0.0
)

replace gocommerce => ../..
