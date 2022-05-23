# フロントエンド(Typescript or HTML, CSS)
    デザイン系のこと
    作業を進めながらできるところまでする
    とりあえず案１を実装

## 案１：
    ### typescript/simple-webmercari/ src/components/Listing/Listing.tsxの部分(灰色部分)に「詳細」と「傷状態」
        [] テキストで状態を入力する欄を作る
        [] チェック印をlist the itemの前に設置して、そこにチェックが入っていれば、画像処理で状態を測る処理を実行する。

    ### typescript/simple-webmercari/ src/components/ItemList/ItemList.tsx(表示部分)に「詳細」と「傷状態」
        []すると画像の下のName, Categoryに加えて, Conditionが表示される　。
    
    ### それぞれのファイルの下の方のhtmlっぽいところで見た目を編集、「ミドルウェア」はその上のJSで各データをmain.pyに送る部分
        []OpenCVによって傷検知を行って状態を割り当てた商品には「AI査定済み」のようなバッチをつける？？



## 案２：
    実際のアプリにちかい出品ボタンを実装して、それを押すと詳細入力画面に飛び、入力、出品すると商品一覧に戻る
    ??Typescriptに拘らずHTML, CSSでフロントエンドを実装しても良いのでは
    
→初めは案１で実装して余裕があれな案２にする
