# STEP4: Run the application in a virtual environment

In this step, we will learn how to use Docker.

**:book: Reference**

* [docker docs](https://docs.docker.com/get-started/overview/)
* [Udemy - ゼロからはじめる Dockerによるアプリケーション実行環境構築](https://www.udemy.com/course/docker-k/)

## 1. Install Docker
**Install docker of the latest version, and check if you can run `docker -v`.**

**:book: Reference**

* [Download and Install Docker](https://docs.docker.com/get-docker/)


## 2. Run Docker commands
**Make sure that you're in `mercari-build-training-2022/` directory, and run the following command.**

```shell
$ docker run -v $(pwd)/data/text_en.png:/tmp/img.png wakanapo/tesseract-ocr tesseract /tmp/img.png stdout -l eng
```

What message was diplayed after running this command?

Running this command downlods the corresponding docker image from [the registry](https://hub.docker.com/repository/docker/wakanapo/tesseract-ocr) to your local machine.

This docker image has a functionality to read texts from images (OCR).
Using a docker allows you to run applications using an environment built within the docker image without altering your local system.

```shell
$ docker run -v $(pwd)/data/text_ja.png:/tmp/img.png wakanapo/tesseract-ocr tesseract /tmp/img.png stdout -l jpn
```

**Check if the texts are correctly picked up using any image of your choice containing English or Japanese texts.**

**:beginner: Points**

* Make sure you understand [docker volume](https://docs.docker.com/storage/volumes/) 

## 3. Get Docker Image

**Run the following command.**
```shell
$ docker images
```
This command shows the list of images existing on your local host.
You can see that the image we used in the previous step called `wakanapo/tesseract-ocr` is listed here.

**Run the following command and see different types of Docker commands**
```
$ docker help
```
Docker will download images automatically if they are not found on your local system. You can also download the image beforehand.


**Look for a commmand to download an image from the registry and download an image called `alpine`**

Check that you can see `alpine` in the list of images.

**:book: Reference**

* [Docker commands](https://docs.docker.com/engine/reference/commandline/docker/)

**:beginner: Points**

Make sure you understand the following commands and when to use them.

* images
* help
* pull


## 4. Building a Docker Image
**Build the docker file under the directory `python/` if you're using Python and `go/` if you're using Go.**

* Set the name of the image to be `build2022/app` with `latest` tag.

Check that you can now see `build2022/app` in the list of images.


**:book: Reference**

* [Dockerfile reference](https://docs.docker.com/engine/reference/builder/)

## 5. Modity Dockerfile
**Run the docker image you built in STEP4-3, and check if the following error shows up.**

```
docker: Error response from daemon: OCI runtime create failed: container_linux.go:380: starting container process caused: exec: "python": executable file not found in $PATH: unknown.
ERRO[0000] error waiting for container: context canceled 
```

`"python"` part will be replaced with `"go"` if you're using Go.


**Modify the `dockerfile` so that you can use the same version of Python/Go as STEP2 in your docker image.**

Run the image with the modified `dockerfile`, check if the same message is displayed as STEP2-2.

**:book: Reference**

* [docker docs - language guide (Python)](https://docs.docker.com/language/python/)
* [docker docs - language guide (Go)](https://docs.docker.com/language/golang/)

## 6. Run the listing API on Docker

The environment within the docker image should be the same as STEP2-2 after STEP4-4.

**Mofify `dockerfile` to copy necessary files and install dependencies such that you can run the listing API on docker**


`$ docker run -d -p 9000:9000 build2022/app:latest`

Check if the above command results in the same response as STEP3.

---
**:beginner: Points**

Make sure you understand the following concepts

* images
* pull
* build
* run
* Dockerfile

---

### Next

[STEP5: Implement a simple Mercari webapp as frontend](step5.en.md)