# STEP5: Run the application in a virtual environment

## 2. Run Docker commands

## 3. Get Docker Image

* Make sure you understand the following commands and when to use them.
A: Here areSome common Docker comands.
`docker –version`
Get the currently installed version of docker.

`docker pull <image_name>`
Pull images from the docker repository.

`docker run -it -d <image_name>`
Create a container from an image.

`docker images`
List all the locally stored docker images

`docker rmi -f <image_id>`(very useful)
delete unnecessary images

## 5. Modify Dockerfile

After modifying the dockerfile, run
`docker build -t build2024/app:latest -f python/Dockerfile .` (don't forget the dot)
to create the image, and run
`docker run build2024/app:latest`
and the output would be
`Python 3.9.18`

## 6. Run the listing API on Docker

* Here's the memo of the first git:
I got a problem here.
`python/main.py` works well on the local host.
However, after I run
`docker run build2024/app:latest`
When the api is initializing with the function `start_connection`,
there's some error in row 76 "unable to open database file".
I think that the path may be wrong. But I have no idea how to check the file structure in the docker.

* Solution

To check the file structure in the block by
`dive <image_id or image_name>`

For example, after creating the image `build2024/app:latest`, roll to the Filetree app and check it.

What's more, after executing `CMD` in `dockerfile`, the path is where `python.py` is. Thus, the original path code
`path = pathlib.Path(__file__).parent.resolve()`
is changed into
`path = pathlib.Path(__file__).parent.parent.resolve()`