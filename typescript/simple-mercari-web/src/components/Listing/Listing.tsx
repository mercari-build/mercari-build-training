import React, { useState } from 'react';

const server = process.env.API_URL || 'http://127.0.0.1:9000';

interface Prop {
  onListingCompleted?: () => void;
}

type formDataType = {
  name: string,
  category: string,
  image: string | File,
}

export const Listing: React.FC<Prop> = (props) => {
  const { onListingCompleted } = props;
  const initialState = {
    name: "",
    category: "",
    image: "",
  };
  const [values, setValues] = useState<formDataType>(initialState);

  const onNameChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values, [event.target.name]: event.target.value,
    })
  };

  // This function is triggered when the select changes
  // const selectChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
  //   // const value = event.target.value;
  //   // setSelectedOption(value);
  //   setState({
  //     ...values, [event.target.name]: event.target.value,
  //   })
  // };

  const onFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values, [event.target.name]: event.target.files![0],
    });
  };
  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const data = new FormData()
    data.append('name', values.name)
    data.append('category', values.category)
    data.append('image', values.image)

    fetch(server.concat('/items'), {
      method: 'POST',
      mode: 'cors',
      body: data,
    })
      .then(response => response.json())
      .then(data => {
        console.log('POST success:', data);
        onListingCompleted && onListingCompleted();

        // clear input after submission
        setValues(initialState);
        (document.getElementById('imageName') as HTMLInputElement).value = "";
      })
      .catch((error) => {
        console.error('POST error:', error);
      })
  };
  return (
    <div>
      <form onSubmit={onSubmit}>
        <div className='Listing'>
          <input type='text' name='name' id='nameInput' placeholder='Name' value={values.name} onChange={onNameChange} required />
          {/* potentially make option @e-yang08*/}
          <input type='text' name='category' id='categoryInput' placeholder='Category' value={values.category} onChange={onNameChange} required />
          {/* <select name='category' id='categoryInput'  required>
            <option selected disabled hidden className='placeHolder'>category</option>
            <option value='Fashion'>Fashion</option>
            <option value='Beauty'>Beauty</option>
            <option value='Home'>Home</option>
            <option value='Furniture'>Furniture</option>
            <option value='Jewelry'>Jewelry</option>
            <option value='Kids'>Kids</option>
            <option value='Toys'>Toys</option>
            <option value='Books'>Books</option>
          </select> */}
          {/* potentially make option @e-yang08 */}
          <input type='text' name='brand' id='brandInput' placeholder='Brand' value={values.name} onChange={onNameChange} required />
          {/* potentially make option @e-yang08 */}
          <input type='text' name='size' id='sizeInput' placeholder='Size' value={values.name} onChange={onNameChange} required />
          <input type='text' name='product' id='productNumInput' placeholder='Product # (Optional)' value={values.name} onChange={onNameChange} />
          <textarea name='details' id='detailsInput' cols={30} rows={5} placeholder="Enter text"></textarea>
          <input type='file' name='image' id='imageInput' onChange={onFileChange} />
          <label className='upload-btn' htmlFor='imageInput'>
            <img id='upload-btn-img' src='upload-icon.png' alt='upload icon'></img>
            Image
          </label>
          <button type='submit' id='submit-button'>List this item</button>
        </div>
      </form>
    </div>
  );
}