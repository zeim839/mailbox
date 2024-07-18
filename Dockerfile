# syntax=docker/dockerfile:1
FROM golang:1.22

WORKDIR /app

# Copy module definitions and install dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy submodules, config environment, server source.
COPY config ./config/
COPY core ./core/
COPY data ./data/
COPY server/*.go config.env ./

# Compile GoLang.
RUN CGO_ENABLED=0 GOOS=linux go build -o /mailbox

# Expose server port.
EXPOSE 8080

# Run server.
CMD ["/mailbox"]
