import React, { useState } from 'react';

const server = process.env.API_URL || 'http://127.0.0.1:9000';

interface Prop {
  onListingCompleted?: () => void;
}

type formDataType = {
  name: string,
  category: string,
  image_filename: string | File,
}

export const Listing: React.FC<Prop> = (props) => {
  const { onListingCompleted } = props;
  const initialState = {
    name: "",
    category: "",
    image_filename: "",
  };
  const [values, setValues] = useState<formDataType>(initialState);

  const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values, [event.target.name]: event.target.value,
    })
  };
  const onFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values, [event.target.name]: event.target.files![0],
    })
  };

  const clear =  () =>{
    setValues(initialState)
  }

  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const data = new FormData()
    data.append('name', values.name)
    data.append('category', values.category)
    data.append('image_filename', values.image_filename)
    fetch(server.concat('/items'), {
      method: 'POST',
      mode: 'cors',
      body: data,
    })
      .then(response => {
        console.log('POST status:', response.statusText);
        onListingCompleted && onListingCompleted();
      })
      .catch((error) => {
        console.error('POST error:', error);
      })
      clear();
  };

  return (
    <div className='Listing'>
      <form onSubmit={onSubmit}>
        <div className='ListingForm'>
          <input type='text' name='name' id='name' placeholder='name' value={values.name} onChange={onValueChange} required />
          <input type='text' name='category' id='category' placeholder='category' value={values.category} onChange={onValueChange} />
          <input type='file' name='image_filename' id='image_filename' onChange={onFileChange} required />
          <button className='btn' type='submit'>List this item</button>
        </div>
      </form>
    </div>
  );
}
