import React, { useState } from 'react';

const server = process.env.API_URL || 'http://127.0.0.1:9000';

interface Prop {
  onListingCompleted?: () => void;
}

type formDataType = {
  name: string,
  category: string,
  image: string | File,
  condition: string,
  damage_analysis: string,
}

export const Listing: React.FC<Prop> = (props) => {
  const { onListingCompleted } = props;
  const initialState = {
    name: "",
    category: "",
    image: "",
    condition: "",
    damage_analysis: "",
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
  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    const data = new FormData()
    data.append('name', values.name)
    data.append('category', values.category)
    data.append('image', values.image)
    data.append('condition', values.condition)
    data.append('damage_analysis', values.damage_analysis)

    // To test what values are sent over
    // console.log(values.condition)
    // console.log(values.damage_analysis)

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

  };
  return (
    <div className='Listing'>
      <form onSubmit={onSubmit} className="form">
          <div className="group"> 
            <input className="form_input" type='text' name='name' id='name' placeholder='Name' onChange={onValueChange} required />
            <input className="form_input" type='text' name='category' id='category' placeholder='Category' onChange={onValueChange} />
            <input className="form_file" type='file' name='image' id='image' onChange={onFileChange} required />
            <br />
            <input className="form_input" type="text" name="condition" id='condition' placeholder='Describe the item condition' onChange={onValueChange} required />
            <label><input type="checkbox"  name="damage_analysis" id="damage_analysis" onChange={onValueChange}/>Analyse the condition</label>
            <button className="form_submit" type='submit'>List this item</button>
          </div>
      </form>
    </div>
  );
}
