# STEP8: Run frontend and API using docker-compose

## 1. (Revision) Building Docker Images

Clean install npm:
```shell
npm ci
```

After executing:
```shell
npm start
```
`npm` would find `package.json` and execute `start` in `scripts` parts, which is `react-scripts start` here. It is used to run React applications in my local development environment.

## 2. Installing Docker Compose

Docker Desktop includes Docker Compose.
```shell
$ docker-compose -v
```
Output:
```shell
Docker Compose version v2.24.5-desktop.1
```


## 3. Docker Compose Tutorial

* How many services are defined in the docker-compose file in the tutorial? What exactly do these services do?
> A `web` service uses an image that's built from the Dockerfile in the current directory. It then binds the container and the host machine to the exposed port, 8000. This example service uses the default port for the Flask web server, 5000.
> The `redis` service uses a public Redis image pulled from the Docker Hub registry.
(Copy from references)
```shell
services:
  web:
    build: .
    ports:
      - "8000:5000"
  redis:
    image: "redis:alpine"
```

* web service and redis services get docker images with different methods. When running `docker-compose up`, check how where each image id downloaded.
> To list local images,
```shell
docker image ls
```

* In docker-compose, you can connect to different services from a service. How does the web service resolve the name for the redis service and connect to it?

## 4. Run frontend and API using Docker Compose

In `App.css`, change `.ItemList{}` part.