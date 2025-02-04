import { useState } from 'react';
import { postItem } from '../api';

interface Prop {
  onListingCompleted?: () => void;
}

type formDataType = {
  name: string;
  category: string;
  image: string | File;
};

export const Listing: React.FC<Prop> = (props) => {
  const { onListingCompleted } = props;
  const initialState = {
    name: '',
    category: '',
    image: '',
  };
  const [values, setValues] = useState<formDataType>(initialState);

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
    postItem({
      name: values.name,
      category: values.category,
      image: values.image,
    })
      .then((response) => {
        console.log('POST status:', response.statusText);
        if (onListingCompleted) {
          onListingCompleted();
        }
      })
      .catch((error) => {
        console.error('POST error:', error);
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
            onChange={onValueChange}
            required
          />
          <input
            type="text"
            name="category"
            id="category"
            placeholder="category"
            onChange={onValueChange}
          />
          <input
            type="file"
            name="image"
            id="image"
            onChange={onFileChange}
            required
          />
          <button type="submit">List this item</button>
        </div>
      </form>
    </div>
  );
};
