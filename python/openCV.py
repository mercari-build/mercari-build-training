import cv2
#import random
import numpy as np

# 画像を読み込む
#　画像読み込みの方法がわかったらこの行は削除する
img = cv2.imread('../image/cellphone_orig.jpg', 0)

def condition(img):

    # 元画像を読み込み、コントラストを調整する
    img = cv2.medianBlur(img,5)
    img_adpth = cv2.adaptiveThreshold(img,255,cv2.ADAPTIVE_THRESH_MEAN_C,\
                cv2.THRESH_BINARY,31,5)
    cv2.imwrite('../image/cellphone_adpth2.jpg', img_adpth)


    # スマートフォンの外部を真っ黒にするフィルターを作る
    ret,img_bin = cv2.threshold(img,110,255,cv2.THRESH_BINARY)
    img_bin = cv2.bitwise_not(img_bin) # 白黒反転
    # カーネルを作成
    kernel = cv2.getStructuringElement(cv2.MORPH_ELLIPSE, (5, 5))
    # 膨張
    img_bin = cv2.morphologyEx(img_bin, cv2.MORPH_OPEN, kernel, iterations=3)
    # 収縮
    img_bin = cv2.morphologyEx(img_bin, cv2.MORPH_CLOSE, kernel, iterations=20)
    cv2.imwrite('../image/cellphone_mask.jpg', img_bin)
    img_obj = cv2.imread('../image/cellphone_adpth2.jpg', 0)


    # コントラスを調整した元画像に、スマートフォンの外部と内部を区別するフィルターをかける
    # これを"前処理画像"とする
    img_mask = cv2.imread('../image/cellphone_mask.jpg', 0)
    img_mask_inv = cv2.bitwise_not(img_mask)
    img_obj = cv2.bitwise_or(img_obj, img_mask_inv)
    img_obj = cv2.bitwise_not(img_obj)
    cv2.imwrite('../image/cellphone_obj.jpg', img_obj)

    # 前処理画像から赤枠を作成し、元画像に重ねる
    # 元画像をグレースケールで読み込み
    img_orig = cv2.imread('../image/cellphone_orig.jpg', 0)
    # 前処理画像をグレースケールで読み込み
    img_gray = cv2.imread('../image/cellphone_obj.jpg', 0)
    # ラベリング処理
    n, label, data, center = cv2.connectedComponentsWithStats(img_gray)
    sizes = data[1:, -1]
    #print(sizes)
    # ラベリング結果書き出し準備
    img_result = cv2.cvtColor(img_orig, cv2.COLOR_GRAY2BGR)
    #スマートフォンの画素数（マスクの画像の白い部分）
    white_area=cv2.countNonZero(img_bin)
    # サイズフィルター
    for i in range(1, n):
        if 100<sizes[i-1] and sizes[i-1]<(white_area/1000):
            #img_result[label==i] = 255
            # 各オブジェクトの外接矩形を赤枠で表示
            x0 = data[i][0]-2
            y0 = data[i][1]-2
            x1 = x0 + data[i][2] +4
            y1 = y0 + data[i][3] +4
            cv2.rectangle(img_result, (x0, y0), (x1, y1), color=(0, 0, 255), thickness=3)
    # 結果画像を書き出し
    cv2.imwrite('../image/cellphone_recog_result.jpg', img_result)



    #スマートフォンの画素数の1000分の1未満1未満、100ピクセル以上の,傷をカウントする
    sizes = sizes[sizes < (white_area/1000)]
    sizes = sizes[100 < sizes]

    #傷のスコア（スマートフォンの面積に対する、赤枠の総面積の割合, 0〜1）
    scr_score = sum(sizes)/white_area

    return scr_score
