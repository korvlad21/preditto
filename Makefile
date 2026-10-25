SWAG_VERSION := v1.16.6
SWAG := go run github.com/swaggo/swag/cmd/swag@$(SWAG_VERSION)
# Scan package directories (internal itself has no Go files). -e allows bootstrapping
# before the generated package exists; swag still fails on unresolved schemas.
SWAG_DIRS = cmd/api,$(shell cd backend && go list -e -mod=readonly -f '{{.ImportPath}}' ./internal/... | sed 's|^preditto/||' | paste -sd, -)
SWAG_FLAGS = init --generalInfo main.go --dir $(SWAG_DIRS) --parseInternal --outputTypes go --packageName swagger --generatedTime=false

.PHONY: swagger swagger-check
swagger:
	cd backend && $(SWAG) $(SWAG_FLAGS) --output docs/swagger

# Generate outside the checkout: detects drift even before the output is committed.
swagger-check:
	@set -eu; output=$$(mktemp -d); trap 'rm -rf "$$output"' EXIT; \
	cd backend; $(SWAG) $(SWAG_FLAGS) --output "$$output"; \
	diff -u docs/swagger/docs.go "$$output/docs.go"
	cd backend && go test -mod=readonly ./cmd/api -run Swagger -count=1
