# STEP5: Run the application in a virtual environment

## 2. Run Docker commands

## 3. Get Docker Image

* Make sure you understand the following commands and when to use them.
A: Here areSome common Docker comands.
`docker –version`
Get the currently installed version of docker.

`docker pull <image name>`
Pull images from the docker repository.

`docker run -it -d <image name>`
Create a container from an image.

`docker images`
List all the locally stored docker images

## 5. Modify Dockerfile

After modifying the dockerfile, run
`docker build -t build2024/app:latest .`
to create the image, and run
`docker run build2024/app:latest`
and the output would be
`Python 3.9.18`

## 6. Run the listing API on Docker

I got a problem here.
`python/main.py` works well on the local host.
However, after I run
`docker run build2024/app:latest`
When the api is initializing with the function `start_connection`,
there's some error in row 76 "unable to open database file".
I think that the path may be wrong. But I have no idea how to check the file structure in the docker.