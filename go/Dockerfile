FROM golang:1.24.0-alpine

# install packages 
RUN apk add --no-cache git gcc musl-dev sqlite-dev

# set working directory
WORKDIR /app

# download go module
COPY go.mod go.sum ./
RUN go mod download

# copy all source code 
COPY . ./

# create user and group, and set the permission
RUN addgroup -S mercari && adduser -S trainee -G mercari \
    && chown -R trainee:mercari /app/db /app/images


# make images directory and set the owner 
RUN mkdir -p /app/images && \
    chown -R trainee:mercari /app/db /app/images

# build main.go
RUN go build -o server ./cmd/api/main.go

#  set the path to items.db　and images (setting environment variables)
ENV DB_PATH=/app/db/items.db
ENV IMAGE_DIR=/app/images


# set the execution user in the container
USER trainee

# launch the server 
CMD ["./server"]