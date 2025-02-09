# STEP7: Web のフロントエンドを実装する

## 1. 環境構築

以下から v20 の Node をインストールしてください。
（2025 年 2 月現在 v22.13.1 LTS を推奨）

https://nodejs.org/en/

複数のバージョンをインストールしたい場合は[nvs](https://github.com/jasongin/nvs)を推奨します。

`node -v` を実行して `v22.0.0` 以上のバージョンが表示されれば正しくインストールできています。

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

これらは、それぞれ`src/components/Listing.tsx`と`src/components/ItemList.tsx`というコンポーネントによって作られており、`App.tsx`から呼び出されています。

:pushpin: サンプルコードは React で書かれていますが、React の理解は必須ではありません。

### (Optional) 課題 1. 新しい商品を登録する

Listing のフォームを使って、新しい商品を登録してみましょう。この画面では、名前、カテゴリ、画像が登録できるようになっています。

STEP3 で名前とカテゴリのみで出品をする API を作った人は、`typescript/simple-mercari-web/src/components/Listing.tsx`を編集して画像のフィールドを削除しておきましょう。

### (Optional) 課題 2. 各アイテムの画像を表示する

この画面では、商品の画像が Build@Mercari のロゴになっています。`http://localhost:9000/image/<item_id>.jpg`を画像として指定し、一覧画面でそれぞれの画像を表示してみましょう。

### (Optional) 課題 3. HTML と CSS を使ってスタイルを変更する

この二つのコンポーネントのスタイルは、CSS によって管理されています。

どのような変更が加えられるのかを確認するため、まず`ItemList`コンポーネントの CSS を編集してみましょう。`App.css`の中に、この二つのコンポーネントのスタイルが指定されています。ここで指定したスタイルは、`className`という attribute を通じて適用することができます(e.g. `<div className='Listing'></div>`)。

```css
.Listing {
  ...;
}
.ItemList {
  ...;
}
```

CSS だけではなく、各コンポーネントで return されている HTML タグも変更してみましょう。

### (Optional) 課題 4. アイテム一覧の UI を変更する

現在の`ItemList`では、それぞれのアイテムが上から一つずつ表示されています。以下のレファレンスを参考に、グリッドを使ってアイテムを表示してみましょう。

**:book: References**

- [HTML の基本](https://developer.mozilla.org/ja/docs/Learn/Getting_started_with_the_web/HTML_basics)

- [CSS の基本](https://developer.mozilla.org/ja/docs/Learn/Getting_started_with_the_web/CSS_basics)

- [グリッドレイアウトの基本概念](https://developer.mozilla.org/ja/docs/Web/CSS/CSS_Grid_Layout/Basic_Concepts_of_Grid_Layout)

---

### Next

[STEP8: docker-compose で API とフロントエンドを動かす](08-docker-compose.ja.md)
