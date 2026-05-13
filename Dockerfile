FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /api cmd/api/main.go

FROM gcr.io/distroless/static-debian12
COPY --from=builder /api /api
EXPOSE 3000
ENTRYPOINT ["/api"]
