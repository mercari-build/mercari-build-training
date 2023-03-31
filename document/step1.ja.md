# STEP1: Git

このステップではGitとGithubの使い方を学びます。

**:book: Reference**
* [Gitを使ったバージョン管理](https://backlog.com/ja/git-tutorial/intro/01/)
* [Udemy - Git & Github基礎講座](https://www.udemy.com/course/git-github-a/)

## **mercari-build-training-2022** リポジトリをフォークする

* [mercari-build-training-2022](https://github.com/mercari-build/mercari-build-training-2022)
  をあなたのGithubにForkします。
* Forkに成功すると `https://github.com/<your github id>/mercari-build-training-2022`
  というようなリポジトリができます。

## Gitをインストールする
1. Gitをご自身のPCにインストールしてください。以下のコマンドが動けばOKです。
   ```shell
   $ git version
   ```

   * Macを使っている場合: [brew](https://brew.sh/index_ja) をインストールしてから `brew install git`を実行
   * For Windows user: Download [installer](https://gitforwindows.org/)

2. git configに自分の名前とemailアドレスを設定します。以下のコマンドを実行してあなたのemailアドレスが表示されればOKです。
   ```shell
   $ git config user.email
   <your-email-adress>
   ```
   
## Gitの基本コマンドを使う

1. `https://github.com/<your github id>/mercari-build-training-2022` を **clone**
   します。 cloneすると、github上のリポジトリを自分のローカルにDownloadできます。
   ```shell
   $ cd <your working space>
   $ git clone https://github.com/<your github id>/mercari-build-training-2022
   ```

**:bangbang: 注意**

cloneができたら必ず以下のコマンドを実行してください。
```shell
$ cd mercari-build-training-2022
$ git config --local core.hooksPath .githooks/ 
```
これは mercari-build-training-2022 が githooks という機能を使うために必要なものです。

2. `first-pull-request`というブランチを作り、そのブランチに**checkout**します
   ```shell
   $ cd <your working space>/mercari-build-training-2022
   $ git branch first-pull-request
   $ git checkout first-pull-request
   ```
3. README.md の中にある`@<your github id>` の部分をあなたのgithub idに書き換えてください
4. 書き換えた内容を **commit**します
   ```shell
   $ git status # Check your change
   $ git add README.md # README.mdの変更をcommit対象にする
   $ git commit -m "Update github id" # どんな変更を加えたのかを伝えるコメント
   ```
5. 変更内容をgithubに**push**します
   ```shell
   $ git push origin first-pull-request:first-pull-request
   ```
6. `https://github.com/<your github id>/mercari-build-training-2022`を開き、**Pull Request**(PR)を作ります。
    - base branch: `main`
    - target branch: `first-pull-request`

## PRのレビューをする、PRのレビューをもらう
- PRができたら、チームメイトにそのPRのURLを見てもらいます
- 1人以上に`approve`をもらえたらそのPRをmainブランチにmergeします
- また、チームメイトのPRを開いて **変更内容を確認**し、`approve` しましょう。

---

**:book: Reference**
- [コードレビューの仕方](https://fujiharuka.github.io/google-eng-practices-ja/ja/review/reviewer/)

**:beginner: Point**

以下のキーワードについて理解できているか確認しましょう。

- branch
- commit
- add
- pull, push
- Pull Request

---
### Next

[STEP2: 環境構築](step2.ja.md)