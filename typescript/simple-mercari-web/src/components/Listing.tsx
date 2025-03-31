import { useRef, useState } from 'react';
import { postItem } from '~/api';

interface Prop {
  onListingCompleted: () => void;
}

type FormDataType = {
  name: string;
  category: string;
  image: string | File;
};

export const Listing = ({ onListingCompleted }: Prop) => {
  const initialState = {
    name: '',
    category: '',
    image: '',
  };
  const [values, setValues] = useState<FormDataType>(initialState);

  const uploadImageRef = useRef<HTMLInputElement>(null);

  const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values,
      [event.target.name]: event.target.value,
    });
  };
  const onFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values,
      [event.target.name]: event.target.files![0],
    });
  };
  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    console.log("🚀 Form submitted!");  // ✅ デバッグログ

    console.log("🔍 Current state before sending:", values); // ✅ 追加: 送信前の `values` の中身を確認

  if (!values.name.trim() || !values.category.trim()) {
    console.error("❌ Name or category is empty!");
    alert("Name and category are required!");
    return;
  }

  if (!values.image || typeof values.image === "string") {
    console.error("❌ Image is missing or invalid!");
    alert("Please select an image file!");
    return;
  }
  

    postItem({
      name: values.name,
      category: values.category,
      image: values.image,
    })
      .then(() => {
        alert('Item listed successfully');
      })
      .catch((error) => {
        console.error('POST error:', error);
        alert('Failed to list this item');
      })
      .finally(() => {
        onListingCompleted();
        setValues(initialState);
        if (uploadImageRef.current) {
          uploadImageRef.current.value = '';
        }
      });
  };
  return (
    <div className="Listing">
      <form onSubmit={onSubmit}>
        <div>
          <input
            type="text"
            name="name"
            id="name"
            placeholder="name"
            value={values.name} // ✅ 追加
            onChange={onValueChange}
            required
          />
          <input
            type="text"
            name="category"
            id="category"
            placeholder="category"
            value={values.category} // ✅ 追加
            onChange={onValueChange}
          />
          <input
            type="file"
            name="image"
            id="image"
            onChange={onFileChange}
            required
            ref={uploadImageRef}
          />
          <button type="submit">List this item</button>
        </div>
      </form>
    </div>
  );
};
