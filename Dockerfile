# Build stage
FROM golang:alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
RUN CGO_ENABLED=0 go build -o /app/vxlan ./main.go

# Runtime stage
FROM alpine:3.19

# Install iproute2 and iputils to aid in debugging / verifying the network setup
RUN apk add --no-cache iproute2 iputils

COPY --from=builder /app/vxlan /usr/local/bin/vxlan

CMD ["/usr/local/bin/vxlan"]
