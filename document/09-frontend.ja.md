# STEP9: Web のフロントエンドを実装する

## 1. 環境構築

以下から v22 の Node をインストールしてください。
（2025 年 2 月現在 v22.13.1 LTS を推奨）

https://nodejs.org/en/

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

STEP4 のサーバー(Python/Go)もローカルで立ち上げておきましょう。
このシンプルな画面では、以下の二つのことができるようになっています。

- 新しい商品の登録 (Listing)
- 商品の一覧の閲覧 (ItemList)

これらは、それぞれ`src/components/Listing.tsx`と`src/components/ItemList.tsx`というコンポーネントによって作られており、`App.tsx`から呼び出されています。

:pushpin: サンプルコードは [React](https://react.dev/versions) で書かれていますが、React の理解は必須ではありません。なお、ビルドツールとして [React の公式ドキュメントの推奨](https://react.dev/learn/building-a-react-framework#step-1-install-a-build-tool)に従い、[Vite](https://github.com/vitejs/vite) を採用しています。

## (Optional) 課題 1. 新しい商品を登録する

Listing のフォームを使って、新しい商品を登録してみましょう。この画面では、名前、カテゴリ、画像が登録できるようになっています。

STEP4 で名前とカテゴリのみで出品をする API を作った人は、`typescript/simple-mercari-web/src/components/Listing.tsx`を編集して画像のフィールドを削除しておきましょう。

## (Optional) 課題 2. 各アイテムの画像を表示する

この画面では、商品の画像が Build@Mercari のロゴになっています。`http://localhost:9000/images/<item_id>.jpg`を画像として指定し、一覧画面でそれぞれの画像を表示してみましょう。

## (Optional) 課題 3. HTML と CSS を使ってスタイルを変更する

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

## (Optional) 課題 4. アイテム一覧の UI を変更する

現在の`ItemList`では、それぞれのアイテムが上から一つずつ表示されています。以下のドキュメントを参考に、グリッドを使ってアイテムを表示してみましょう。

**:book: References**

- [HTML の基本](https://developer.mozilla.org/ja/docs/Learn/Getting_started_with_the_web/HTML_basics)

- [CSS の基本](https://developer.mozilla.org/ja/docs/Learn/Getting_started_with_the_web/CSS_basics)

- [グリッドレイアウトの基本概念](https://developer.mozilla.org/ja/docs/Web/CSS/CSS_Grid_Layout/Basic_Concepts_of_Grid_Layout)

---

## Tips

### デバッグ

デバッグとは、プログラムの動作を確認し、問題を特定・修正するプロセスです。

Web フロントエンドでは、コードの動作を確認したい箇所に`console.debug()`を仕込むことで、実行時の値や状態を確認することができます。例えば`ItemList.tsx`の場合：

```typescript
export const ItemList = (props: Prop) => {
  ...
  useEffect(() => {
    const fetchData = () => {
      fetchItems()
        .then((data) => {
          console.debug('GET success:', data); // ここでAPIから取得されたデータの中身を確認
          ...
        })
        .catch((error) => {
          console.error('GET error:', error);
        });
    };
  ...
```

これらのデバッグ情報は、ブラウザの開発者ツール（**Chrome DevTools**）で確認することができます。Chrome DevTools は以下のいずれかの方法で開くことができます：

- キーボードショートカット:
  - Windows/Linux: `Ctrl + Shift + I`
  - macOS: `Cmd + Option + I`
- ブラウザ上で右クリックして「検証」を選択
- メニューから「その他のツール」>「デベロッパー ツール」を選択

開発者ツールの「Console」タブに、`console.debug()`で出力した情報が表示されます。

詳しい開発者ツールの使い方については、[Chrome DevTools のドキュメント](https://developer.chrome.com/docs/devtools/open?hl=ja)を参照してください。

### Build Production-Ready App by using Framework

本教材では React の基本的な理解を目的としているため、特定のフレームワークを使用していません。しかし、React の開発チームは実際のプロダクションレベルの Web サービスを開発する際には、以下のようなフレームワークの利用を推奨しています([Creating a React App](https://react.dev/learn/creating-a-react-app))：

- [Next.js (App Router)](https://nextjs.org/docs)
- [React Router (v7)](https://reactrouter.com/start/framework/installation)

:warning: 注意点として、**多くの React の教材等で紹介されている[`create-react-app`](https://github.com/facebook/create-react-app)は非推奨となる**ことが、2025 年 2 月 14 日に[正式にアナウンス](https://react.dev/blog/2025/02/14/sunsetting-create-react-app)されました。新規プロジェクトでは`create-react-app`は使用せず、上述のような別の方法を検討することを強くお勧めします。

今後、実際に長期的にユーザーに使ってもらうことを想定したサービスの開発を担当する際は、これらのフレームワークの利用を検討しましょう。新しいサービスを作る際は、各フレームワークの特徴と作りたいサービスの要件を理解した上で、適切なフレームワークを選定することが重要です。

フレームワークの選定では、以下のような観点を考慮すると良いでしょう：

- サービスの規模と複雑さ
- パフォーマンス要件
- SEO 要件
- チームの技術スタック
- デプロイ環境

## Next

[STEP10: docker-compose で API とフロントエンドを動かす](./10-docker-compose.ja.md)
