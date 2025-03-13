const SERVER_URL = import.meta.env.VITE_BACKEND_URL || 'http://127.0.0.1:8000';

console.log("SERVER_URL:", SERVER_URL);

export interface Item {
  id: number;
  name: string;
  category: string;
  image_name: string;
}

export interface ItemListResponse {
  items: Item[];
}

export const fetchItems = async (): Promise<ItemListResponse> => {
  const response = await fetch(`${SERVER_URL}/items`, {
    method: 'GET',
    mode: 'cors',
    headers: {
      'Content-Type': 'application/json',
      Accept: 'application/json',
    },
  });

  if (response.status >= 400) {
    throw new Error('Failed to fetch items from the server');
  }
  return response.json();
};

export interface CreateItemInput {
  name: string;
  category: string;
  image: string | File;
}

export const postItem = async (input: CreateItemInput): Promise<Response> => {
  console.log("🚀 postItem() called with input:", input); // ✅ 追加: 関数が呼ばれたか確認

  const data = new FormData();
  data.append('name', input.name);
  data.append('category', input.category);
  data.append('image', input.image);


  console.log("📡 FormData prepared:", [...data.entries()]); // ✅ 追加: FormData の中身を確認

  try {
    console.log(`📡 Sending POST request to: ${SERVER_URL}/items`); // ✅ 追加: 送信先URLを確認

   const response = await fetch(`${SERVER_URL}/items`, {
     method: 'POST',
     mode: 'cors',
     body: data,
   });

   console.log("✅ Fetch executed, awaiting response..."); // ✅ 追加: `fetch()` が実行されたか確認

   if (!response.ok) {
    throw new Error(`HTTP error! Status: ${response.status}`);
  }

  console.log("✅ Response received:", response); // ✅ 追加: レスポンスを確認



  return response;
 } catch (error) {
    console.error("❌ POST error:", error); // ✅ 追加: エラーが出た場合のログ
    throw error;
 }
};