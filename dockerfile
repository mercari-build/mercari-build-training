# syntax=docker/dockerfile:1

FROM golang:1.22.0-alpine

WORKDIR /app

RUN apk add --no-cache gcc musl-dev

COPY db/ ./db/
COPY go/ ./go/

WORKDIR /app/go

RUN go mod tidy
RUN go build -o ./mercari-build-training ./app/*.go

RUN sqlite3 /app/db/mercari.sqlite3 < /app/db/items.db

RUN addgroup -S mercari && adduser -S trainee -G mercari

RUN chown -R trainee:mercari images
RUN chmod -R 755 images
RUN chown -R trainee:mercari ../db
RUN chmod -R 755 ../db

USER trainee

EXPOSE 9000

CMD [ "/app/go/mercari-build-training" ]
