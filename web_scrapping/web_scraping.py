import time
from selenium import webdriver
from selenium.webdriver.common.keys import Keys
from googletrans import Translator
import json

translator = Translator()

# download the chromedriver for the version of chrome browser you are using rn and save the .exe file anywhere you wish to and replace the path with your path
PATH = "C:\\Users\\chaur\Desktop\\chromedriver.exe"

# please keep the json structure like this
# temp.json = {"product_name": " ",
#              "category": " ",  # filled
#              "": " ",
#              "brand": " ",  # filled
#              "product_id": 448784,
#              "size": " ",  # filled
#              "Description": " ",
#              "image": " "}
# product_id = 448784


def scrape():
    with open("temp.json", 'r', encoding='utf-8') as f:
        data = json.load(f)
        product_id = data["product_id"]

    wd = webdriver.Chrome(PATH)
    wd.get("https://www.uniqlo.com/jp/ja/products/E"+str(product_id))

    print(wd.title)
    product_name_jp = wd.title
    product_name_en = translator.translate(
        product_name_jp, src='ja', dest='en')

    # search_btn = wd.find_element_by_class_name("uq-ec-button-icon")
    # search_btn.click()

    # search = wd.find_element_by_id("Search")
    # search.send_keys("item no. "+str(product_id))
    # search.send_keys(Keys.RETURN)

    # # item_box = wd.find_element_by_id(str(product_id)+"-000")
    # item_box = wd.find_element_by_class_name(
    #     "uq-ec-button-toggle uq-ec-button-toggle--with-icon uq-ec-button-toggle--icon-only uq-ec-cursor-pointer uq-ec-product-tile__toggle")
    # item_box.click()
    time.sleep(2)
    # txt = wd.find_element_by_class_name("uq-ec-gutter-container")
    # print("\ntext")
    # print(txt.text)

    summary_btn = wd.find_element_by_id("productLongDescription")
    summary_btn.click()
    # print("\nsummarybtn")
    # print(summary_btn.text)

    details_btn = wd.find_element_by_id("productMaterialDescription")
    details_btn.click()
    # print("\ndetailsbtn")
    # print(details_btn.text)

    time.sleep(2)
    summary = wd.find_element_by_id("productLongDescription-content")
    print("\nsummary")
    print(summary.text)
    summary_jp = summary.text
    summary_en = translator.translate(summary_jp, src='ja', dest='en')
    print(summary_en.text)

    details = wd.find_element_by_id("productMaterialDescription-content")
    print("\ndetails")
    print(details.text)
    details_jp = details.text
    details_en = translator.translate(details_jp, src='ja', dest='en')
    print(details_en.text)

    time.sleep(5)
    wd.quit()

    sizechart = "https://www.uniqlo.com/jp/ja/size/" + \
        str(product_id)+"_size.html"

    data["product_name"] = product_name_jp+" | "+product_name_en.text
    data["Description"] = "日本語-----------\n" + summary_jp + "\n\n" + \
        details_jp + "\nEnglish-----------\n" + \
        summary_en.text + "\n\n" + details_en.text + "\n\nSizechart: " + sizechart

    with open('temp.json', 'w', encoding='utf8') as f:
        json.dump(data, f, indent=4, ensure_ascii=False)
