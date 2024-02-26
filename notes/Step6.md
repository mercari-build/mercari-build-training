# STEP6: Build Docker image using CI

## 1. About CI

Basic concept

A `workflow` is a complete workflow process. Each workflow contains a set of jobs tasks.

A `jobs` task consists of one or more jobs. Each job consists of a series of steps.

Each `step` can either execute a command or use an action.

Each `action` is a generic basic unit.

## 2. Enable GitHub Actions

Github Actions - add a new workflow file
Change the path of the docker file

## 3. Build the application with GitHub Actions and upload the docker image to the registry

Pull the image.
```shell
$ docker pull ghcr.io/dinaelin-yip/mercari-build-training:Step6_CI
```
Run the image.
```shell
$ docker run -d -p 9000:9000 ghcr.io/dinaelin-yip/mercari-build-training:Step6_CI
```
Output:
```shell
INFO:     Uvicorn running on http://0.0.0.0:9000 (Press CTRL+C to quit)
```
Then the API is running successfully:
```shell
$ curl -X POST \
  --url 'http://0.0.0.0:9000/items' \
  -F 'name=sofa' \
  -F 'category=furniture' \
  -F 'image=@/Users/xiaotongye/Desktop/images/sofa.jpg'
```
Output:
```shell
{"message":"item received: sofa"}
```

