# Build stage
FROM golang:1.21 as builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/server ./cmd/server

# Runtime stage
FROM gcr.io/distroless/base-debian12
COPY --from=builder /bin/server /server
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/server"]
