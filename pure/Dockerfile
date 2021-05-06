FROM golang:1.16 AS backend_base
WORKDIR /build
COPY ./go.mod /build
RUN go mod download

FROM backend_base AS backend_builder
COPY ./pkg /build/pkg
COPY ./cmd/server/main.go /build
RUN go build -o /build/app main.go
RUN chmod +x /build/app

FROM backend_base AS backend_tester
COPY ./pkg /build/pkg
RUN go test --race ./...
RUN touch /tmp/ok

FROM golang:1.16
COPY --from=backend_tester /tmp/ok /dev/null
COPY --from=backend_builder /build/app /
ENTRYPOINT ["/app"]