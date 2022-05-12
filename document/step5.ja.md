# STEP5: Webのフロントエンドを実装する

## 1. 環境構築
以下からv16のNodeをインストールしてください。
（2022年5月現在16.15.0 LTSを推奨）

https://nodejs.org/en/

複数のバージョンをインストールしたい場合は[nvs](https://github.com/jasongin/nvs)を推奨します。

`node -v` を実行して `v16.0.0` 以上のバージョンが表示されれば正しくインストールできています。

まずはディレクトリに移動して、必要なライブラリをインストールします。
```shell
cd typescript/simple-mercari-web
npm ci
```

以下のコマンドでアプリを起動させた後、ブラウザから[http://localhost:3000/](http://localhost:3000/)にアクセスします。
```shell
npm start
```
サーバー(Python/Go)もローカルで立ち上げておきましょう。
このシンプルな画面では、以下の二つのことができるようになっています。
- 新しい商品の登録 (Listing)
- 商品の一覧の閲覧 (ItemList)
  
これらは、それぞれ`src/components/Listing`と`src/components/ItemList`というコンポーネントによって作られており、`App.tsx`から呼び出されています。

:pushpin: サンプルコードはReactで書かれていますが、Reactの理解は必須ではありません。

### (Optional) 課題1. 新しい商品を登録する
Listingのフォームを使って、新しい商品を登録してみましょう。この画面では、名前、カテゴリ、画像が登録できるようになっています。

STEP3で名前とカテゴリのみで出品をするAPIを作った人は、`typescript/simple-mercari-web/src/components/Listing/Listing.tsx`を編集して画像のフィールドを削除しておきましょう。


### (Optional) 課題2. 各アイテムの画像を表示する
この画面では、商品の画像がBuild@Mercariのロゴになっています。`http://localhost:9000/image/<item_id>.jpg`を画像として指定し、一覧画面でそれぞれの画像を表示してみましょう。

### (Optional) 課題3. HTMLとCSSを使ってスタイルを変更する
この二つのコンポーネントのスタイルは、CSSによって管理されています。


どのような変更が加えられるのかを確認するため、まず`ItemList`コンポーネントのCSSを編集してみましょう。`App.css`の中に、この二つのコンポーネントのスタイルが指定されています。ここで指定したスタイルは、`className`というattributeを通じて適用することができます(e.g. `<div className='Listing'></div>`)。
```css
.Listing {
  ...
}
.ItemList {
  ...
}
```
CSSだけではなく、各コンポーネントでreturnされているHTMLタグも変更してみましょう。


### (Optional) 課題4. アイテム一覧のUIを変更する
現在の`ItemList`では、それぞれのアイテムが上から一つずつ表示されています。以下のレファレンスを参考に、グリッドを使ってアイテムを表示してみましょう。


**:book: References**

- [HTMLの基本](https://developer.mozilla.org/ja/docs/Learn/Getting_started_with_the_web/HTML_basics)

- [CSSの基本](https://developer.mozilla.org/ja/docs/Learn/Getting_started_with_the_web/CSS_basics)

- [グリッドレイアウトの基本概念](https://developer.mozilla.org/ja/docs/Web/CSS/CSS_Grid_Layout/Basic_Concepts_of_Grid_Layout)

---

### Next

[STEP6: docker-composeでAPIとフロントエンドを動かす](step6.ja.md)