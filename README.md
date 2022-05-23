# Build@Mercari Training Program 2022

This is @<manami-bunbun>'s build training repository.

Build trainingの前半では個人で課題に取り組んでもらい、Web開発の基礎知識をつけていただきます。
ドキュメントには詳細なやり方は記載しません。自身で検索したり、リファレンスを確認したり、チームメイトと協力して各課題をクリアしましょう。

ドキュメントには以下のような記載があるので、課題を進める際に参考にしてください。

In the first half of Build@Mercari program, you will work on individual tasks to understand the basics of web development. Detailed instructions are not given in each step of the program, and you are encouraged to use official documents and external resources, as well as discussing tasks with your teammates and mentors.

The following icons indicate pointers for 

**:book: Reference**

* そのセクションを理解するために参考になるUdemyやサイトのリンクです。課題内容がわからないときにはまずReferenceを確認しましょう。
* Useful links for Udemy courses and external resources. First check those references if you are feeling stuck.

**:beginner: Point**

* そのセクションを理解しているかを確認するための問いです。 次のステップに行く前に、**Point**の問いに答えられるかどうか確認しましょう。
* Basic questions to understand each section. Check if you understand those **Points** before moving on to the next step.

## Tasks

- [x] **STEP1** Git ([JA](document/step1.ja.md)/[EN](document/step1.en.md))
- [x] **STEP2** Setup environment ([JA](document/step2.ja.md)
  /[EN](document/step2.en.md))
- [x] **STEP3** Develop API ([JA](document/step3.ja.md)
  /[EN](document/step3.en.md))
- [x] **STEP4** Docker ([JA](document/step4.ja.md)/[EN](document/step4.en.md))
- [ ] **STEP5** (Stretch) Frontend ([JA](document/step5.ja.md)
  /[EN](document/step5.en.md))
- [x] **STEP6** (Stretch)  Run on docker-compose ([JA](document/step6.ja.md)
  /[EN](document/step6.en.md))

### Other documents

- 効率的に開発できるようになるためのTips / Tips for efficient development ([JA](document/tips.ja.md)/[EN](document/tips.en.md))
	
--- 
	

# Hackweek 5/23-5/31
  
- Kickoff meeting note [ドキュメント(閲覧制限あり)](https://docs.google.com/document/d/12-YTNs6I2TAsNm49sLjNW2BjZ1_bQp2jSE6_KqkZD_Y/edit?userstoinvite=tkat0@mercari.com#)
  
## 注意⚠️
  
- 現在step5にあたる、POSTに対しての画像ファイル処理にエラーが発生しているため、ItemList.tsc内ではPlaceholderファイルを表示している
  	
- ブランチstep5はこのエラー解消用
  
## ブランチ運用(とりあえず25日まで)


	  main—------(念の為おいておく)
				|- HackWeek (デモ用)
					|- frontend(typescript用) :森本さん
					|- Backend(Jupyternoteでopencv処理):大村さん
					|- step5(step5の画像表示エラー解消用):中川
	
	
	
## アプリの動かし方
	
1. 一つ目のターミナルでサーバーでアプリを実行
	
	```
	cd python
	uvicorn main:app --reload --port 9000
	```
	
2. 二つ目のターミナルでフロントエンドを動かす
	
	```
	cd typescript/simple-mercari-web
	```
	
	* ↓一回目だけ　参照([JA](document/step5.ja.md))
	
	```
	npm ci 
	```
	
	* 実行
	```
	npm start
	```

## TODO
  - Frontend :
  - Backend :
  - Error :
  

  
