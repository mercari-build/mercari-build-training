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
```dockerfile
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
$ docker image ls
```

* In docker-compose, you can connect to different services from a service. How does the web service resolve the name for the redis service and connect to it?

## 4. Run frontend and API using Docker Compose

First, reconstruct two images from the very start.

At the root path of the git folder, construct `build2024/app:latest` with `python/dockerfile`.
```shell
$ docker build -t build2024/app:latest -f python/dockerfile .
```

Run it. 
```shell
$ docker run -p 9000:9000 build2024/app:latest
```

Enter directory `typescript/simple-mercari-web`, construct `build2024/web:latest` with `dockerfile`. This is necessary because command `RUN npm install` requires that `package.json` exists.
```shell
$ docker build -t build2024/web:latest -f dockerfile .
```

Run it. No issues.
```shell
$ docker run -p 3000:3000 build2024/web:latest
```

However, in the root directory, execute
```shell
$ docker-compose up
```

I could use `GET` method but I couldn't use `PORT` method to upload items.
When the `PORT` api starts, the debug information is like:
```shell
app-1  | DEBUG:multipart.multipart:Calling on_part_begin with no data
app-1  | DEBUG:multipart.multipart:Calling on_header_field with data[42:61]
...
app-1  | DEBUG:multipart.multipart:Calling on_part_end with no data
app-1  | DEBUG:multipart.multipart:Calling on_end with no data
```
I guess that it may be the problem with path, so I want to print the path. However, after I edit `python/main.py` and rebuild the two docker images once again, `docker-compose up` still runs tha old version of `python/main.py`.
The image `build2024/app:latest` contains 3 folders, `/db`, `/image` and `/python`.
