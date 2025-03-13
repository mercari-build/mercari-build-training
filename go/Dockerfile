FROM golang:1.24.0-alpine

ENV CGO_ENABLED=1 

RUN apk add --no-cache git sqlite-dev gcc musl-dev

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o myapp cmd/api/main.go

RUN addgroup -S mercari && adduser -S trainee -G mercari
RUN chown -R trainee:mercari db

USER trainee

CMD ["/app/myapp"]
