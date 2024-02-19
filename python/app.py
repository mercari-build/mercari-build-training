import json
from flask import Flask, jsonify

app = Flask(__name__)

# JSONファイルのパス
JSON_FILE_PATH = 'items.json'

# JSONファイルからデータを読み込む関数
def read_items_from_json():
    with open(JSON_FILE_PATH, 'r') as file:
        data = json.load(file)
    return data.get('items', [])

# JSONファイルにデータを書き込む関数
def write_items_to_json(items):
    with open(JSON_FILE_PATH, 'w') as file:
        json.dump({'items': items}, file)

@app.route('/items', methods=['GET'])
def get_items():
    items = read_items_from_json()
    return jsonify({"items": items})

if __name__ == '__main__':
    app.run(debug=False)
