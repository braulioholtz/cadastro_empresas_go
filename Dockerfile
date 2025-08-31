# Build stage
FROM golang:1.21 as builder
WORKDIR /src
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /wsserver ./cmd/wsserver
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app ./cmd/runner

# Runtime stage
FROM gcr.io/distroless/base-debian12
COPY --from=builder /server /server
COPY --from=builder /wsserver /wsserver
COPY --from=builder /app /app
EXPOSE 8080 8090
ENV PORT=8080
ENTRYPOINT ["/app"]
