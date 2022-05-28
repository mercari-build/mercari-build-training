import os
import requests
import json
import urllib.parse
from dotenv import load_dotenv
from pathlib import Path

# Rapid API key
dotenv_path = Path('rapidapi_key.env')
load_dotenv(dotenv_path=dotenv_path)
RAPIDAPI_KEY = os.getenv("RAPIDAPI_KEY")

def translate_to_japanese(item_name):
  url = "https://google-translate1.p.rapidapi.com/language/translate/v2"

  payload = "q=" + item_name + "&target=ja&source=en"
  headers = {
    "content-type": "application/x-www-form-urlencoded",
    "Accept-Encoding": "application/gzip",
    "X-RapidAPI-Host": "google-translate1.p.rapidapi.com",
    "X-RapidAPI-Key": RAPIDAPI_KEY
  }

  response = requests.request("POST", url, data=payload, headers=headers)
  json_data = json.loads(response.text)
  return json_data["data"]["translations"][0]["translatedText"]

def translate_to_english(item_name):
  url = "https://google-translate1.p.rapidapi.com/language/translate/v2"

  encoded_item_name = urllib.parse.quote(item_name.encode('utf-8'))

  payload = "q=" + encoded_item_name + "&target=en&source=ja"
  headers = {
    "content-type": "application/x-www-form-urlencoded",
    "Accept-Encoding": "application/gzip",
    "X-RapidAPI-Host": "google-translate1.p.rapidapi.com",
    "X-RapidAPI-Key": RAPIDAPI_KEY
  }

  response = requests.request("POST", url, data=payload, headers=headers)
  json_data = json.loads(response.text)
  return json_data["data"]["translations"][0]["translatedText"]

def detect_language(item_name):
  url = "https://google-translate1.p.rapidapi.com/language/translate/v2/detect"

  payload = "q=" + item_name
  headers = {
    "content-type": "application/x-www-form-urlencoded",
    "Accept-Encoding": "application/gzip",
    "X-RapidAPI-Host": "google-translate1.p.rapidapi.com",
    "X-RapidAPI-Key": RAPIDAPI_KEY
  }

  response = requests.request("POST", url, data=payload, headers=headers)
  json_data = json.loads(response.text)
  return json_data["data"]["detections"][0][0]["language"]