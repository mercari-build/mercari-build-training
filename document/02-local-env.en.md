# STEP2: Building local environment

Choose either Python or Go and build your local environment.

---
## Building Python environment

### 1. Install Python
* If your local version is below Python3.7, install Python3.10.
* If you have Python3.7 or above, you can skip the installation step.

### 2. Check your Python version

* Check if the Python is added to your PATH (usable as commands on your terminal) with the following command.

```shell
$ python -V
```

If the version does not correspond to the Python version you installed, double check your installation as it is not **added to your PATH**.

**:book: Reference**

* [Python - Environment Setup: Setting up PATH](https://www.tutorialspoint.com/python/python_environment.htm)

### 3. Install dependencies

The list of dependent libraries is written in a file called `requirements.txt` in a typical Python project.
You can install the dependencies by running the following command.

```shell
$ cd python
$ pip install -r requirements.txt
```

If you added a library, make sure you add it to `requirements.txt`.

### 4. Run the Python app

```shell
$ uvicorn main:app --reload --port 9000
```

If successful, you can access the local host `http://127.0.0.1:9000` on our browser and you will see`{"message": "Hello, world!"}`.

---

## Building Go environment
### 1. Install Go
* If your local version is below Go1.14, install Go1.18.
* If you have Go1.14 or above, you can skip the installation step.

### 2. Check your Go version

* Check if Go is added to your PATH (usable as commands on your terminal) with the following command.


```shell
$ go version
```

If the version does not correspond to the Go version you installed, double check your installation as it is not **added to your PATH**.

**:book: Reference**

* [GOROOT and GOPATH](https://www.jetbrains.com/help/go/configuring-goroot-and-gopath.html)

### 3. Install dependencies

In Go, dependent libraries are managed in a file called `go.mod`.
You can install the dependencies by running the following command.

```shell
$ cd go
$ go mod tidy
```

**:beginner: Point**

Understand the role of `go.mod` and the commands around it referring to this [document](https://pkg.go.dev/cmd/go#hdr-The_go_mod_file).

### 4. Run the Go app

```shell
$ go run app/main.go
```

If successful, you can access the local host `http://127.0.0.1:9000` on our browser and you will see`{"message": "Hello, world!"}`.

---
**:beginner: Points**

* If you're using Linux or Mac, understand when and how `.bash_profile` and `.bashrc` are activated and used (or `.zshrc` if you're using zsh).
* Understand what it means to **add to PATH**.

**:book: Reference**

The following resources are useful to dive deeper into building environments and Linux.

* (JA)[book - [試して理解]Linuxのしくみ ~実験と図解で学ぶOSとハードウェアの基礎知識](https://www.amazon.co.jp/dp/477419607X/ref=cm_sw_r_tw_dp_178K0A3YTGA97XRH318R)
* (JA)[Udemy Business - もう絶対に忘れない Linux コマンド【Linux 100本ノック+名前の由来+丁寧な解説で、長期記憶に焼き付けろ！](https://mercari.udemy.com/course/linux100test/)
  * ↑わかりやすい講座だと思い貼ってますが、コマンドの暗記は特にしなくていいです

* (EN)[An Introduction to Linux Basics](https://www.digitalocean.com/community/tutorials/an-introduction-to-linux-basics)
* (EN)[Udemy Business - Linux Mastery: Master the Linux Command Line in 11.5 Hours](https://mercari.udemy.com/course/linux-mastery/)
  * You do NOT have to memorize the commands!

---

### Next

[STEP3: Make a listing API](03-api.en.md)
