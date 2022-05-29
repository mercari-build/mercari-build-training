import React, { useState } from 'react';

const server = process.env.API_URL || 'http://127.0.0.1:9000';

export const ListingUser: React.FC<{}> = () => {
    const initialState = {
        name: "",
        password: "",
    };
    const [values, setValues] = useState(initialState);

    const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
        setValues({ ...values, [event.target.name]: event.target.value });
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
        data.append('password', values.password)

        fetch(server.concat('/users'), {
            method: 'POST',
            mode: 'cors',
            body: data,
        })
            .then(response => response.json())
            .then(data => {
                console.log('POST success:', data);
            })
            .catch((error) => {
                console.error('POST error:', error);
            })
    };
    return (
        <div className='ListingUser'>
            <form onSubmit={onSubmit}>
                <div>
                    <input type='text' name='name' id='name' placeholder='name' onChange={onValueChange} required/>
                    <input type='text' name='password' id='password' placeholder='password' onChange={onValueChange} required/>
                    <button type='submit'>login</button>
                </div>
            </form>
        </div>

    );
}
