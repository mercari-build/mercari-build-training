# STEP1: Git

In this step, we will learn how to use Git and Github.

**:book: Reference**
* [Git basics](https://www.atlassian.com/git)
* [GitHub Ultimate: Master Git and GitHub - Beginner to Expert](https://www.udemy.com/course/github-ultimate/)


## Fork this **mercari-build-training-2022** repository

* Fork [mercari-build-training-2022](https://github.com/mercari-build/mercari-build-training-2022) 
* You will see be able to see `https://github.com/<your github id>/mercari-build-training-2022` if successful.


## Install Git
1. Install git in your environment and run the following command.
   ```shell
   $ git version
   ```
   
   * For Mac users: Install [brew](https://brew.sh/index) and `brew install git`
   * For Windows users: Download [installer](https://gitforwindows.org/)
   
2. Set your name and email address using git config. Check if your email address shows up.
   ```shell
   $ git config user.email
   <your-email-adress>
   ```

## Use basic commands in Git

1. **Clone** `https://github.com/<your github id>/mercari-build-training-2022` onto your local using the following command.
   ```shell
   $ cd <your working space>
   $ git clone https://github.com/<your github id>/mercari-build-training-2022
   ```

**:bangbang: Caution**

Please definitely run the following command after cloning repository. 
```
cd mercari-build-training-2022
git config --local core.hooksPath .githooks/ 
```
This is required to use githooks in mercari-build-training-2022 repository.

2. Make a new branch named `first-pull-request` and **checkout** into this branch
   ```shell
   $ cd <your working space>/mercari-build-training-2022
   $ git branch first-pull-request
   $ git checkout first-pull-request
   ```
3. Replace `@<your github id>` on README.md with your Github ID.
4. **commit** the changes you made with the following commands.
   ```shell
   $ git status # Check your change
   $ git add README.md # Add README.md file to the list of files to commit
   $ git commit -m "Update github id" # Brief description about the changes
   ```
5. **push** changes to Github.
   ```shell
   $ git push origin first-pull-request:first-pull-request
   ```
6. Open `https://github.com/<your github id>/mercari-build-training-2022` and make a **Pull Request** (PR).
    - base branch: `main`
    - target branch: `first-pull-request`

## Review a PR and have your PR reviewed
- Once you made a PR, ask a teammate for review.
- If at least one person `approve`s the PR, `merge` into the main branch
- Open your teammates' PRs and check the files changed, and `approve` if you think the changes look good.
---

**:book: Reference**
- [How to do a code review](https://google.github.io/eng-practices/review/reviewer/)


**:beginner: Points**

Check if you understand the following concepts.

- branch
- commit
- add
- pull, push
- Pull Request

---

### Next

[STEP2: Building local environmentSTEP2: Building local environment](step2.en.md)
