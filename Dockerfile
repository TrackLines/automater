FROM golang:1.27-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /automation ./cmd/automation

# runs the night's jobs and exits: no ports, no health check
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /automation /automation
ENTRYPOINT ["/automation"]
