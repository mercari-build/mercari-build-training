# STEP10: Run frontend and API using docker-compose
In this step, we will learn how to use docker-compose.

**:book: Reference**

* (JA)[Docker Compose の概要](https://matsuand.github.io/docs.docker.jp.onthefly/compose/)
* (JA)[Udemy Business - 駆け出しエンジニアのためのDocker入門](https://mercari.udemy.com/course/docker-startup/)

* (EN)[Docker Compose Overview](https://docs.docker.com/compose/)
* (EN)[Udemy Business - Docker for the Absolute Beginner - Hands On - DevOps](https://mercari.udemy.com/course/learn-docker/)

## 1. (Revision) Building Docker Images

**Revisit STEP7 and build a docker image for running the web frontend**

You have a sample `Dockerfile` In `typescript/simple-mercari-web`. Modify this file to run frontend on Docker.

* Set the name of the repository as `mercari-build-training/web` and tag as `latest`.

Run the following and check if you can successfully open [http://localhost:3000/](http://localhost:3000/) on your browser.

`$ docker run -d -p 3000:3000 mercari-build-training/web:latest`


## 2. Installing Docker Compose
**Install Docker Compose and check you can run `docker-compose -v`**

**:book: Reference**

* [Install Docker Compose](https://docs.docker.com/compose/install/)

## 3. Docker Compose Tutorial
**Go through [Docker Compose tutorial](https://docs.docker.com/compose/gettingstarted/)**

:pushpin: Sample code is in Python but the knowledge of Python or the environment is not necessary. Use this tutorial regardless of the backend language you chose in STEP4.

**:beginner: Point**

Let's check if you can answer the following questions.

* How many services are defined in the docker-compose file in the tutorial? What exactly do these services do?
* web service and redis services get docker images with different methods. When running `docker-compose up`, check how where each image id downloaded.
* In docker-compose, you can connect to different services from a service. How does the web service resolve the name for the redis service and connect to it?

## 4. Run frontend and API using Docker Compose
**Referring to the tutorial material, run the frontend and API using Docker Compose**


Set up `docker-compose.yml` under `mercari-build-training/`

Make a new file `docker-compose.yml` considering the following points.

* Docker image to use
    * (Option 1: Difficulty ☆) Use `mercari-build-training/app:latest` and `mercari-build-training/web:latest` made in STEP7 and STEP10-1
    * (Option 2: Difficulty ☆☆☆) Build from `{go|python}/Dockerfile` and `typescript/simple-mercari-web/Dockerfile`
* Port numbers
    * API : 9000
    * Frontend : 3000
* Connecting between services
    * Frontend should send requests to an environment variable `REACT_APP_API_URL`
    * While API will not send requests to frontend, [CORS](https://developer.mozilla.org/ja/docs/Web/HTTP/CORS) needs to be set up such that frontend knows where the requests are coming from
    * Set an environment variable `FRONT_URL` for frontend URL


Run `docker-compose up` and check if the following operates properly
- [http://localhost:3000/](http://localhost:3000/) displays the frontend page
- You can add an new item (Listing)
- You can view the list of all items (ItemList)
