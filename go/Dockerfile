FROM golang:1.24-alpine

RUN apk add --no-cache git gcc musl-dev

# Copy first for better caching on later builds
COPY go.mod go.sum ./
RUN go mod download

WORKDIR /app
COPY . .
RUN go build cmd/api/main.go

# For sqlite3
ENV CGO_ENABLED=1

RUN addgroup -S mercari && adduser -S trainee -G mercari
RUN chown -R trainee:mercari db images

USER trainee

CMD ["./main"]
