import React, {useEffect, useState} from 'react';
import {Listing} from "../Listing";
import {Link} from "react-router-dom";


interface Item {
    id: number;
    name: string;
    category: string;
    image: string;
}

const server = process.env.API_URL || 'http://127.0.0.1:9000';
// const placeholderImage = process.env.PUBLIC_URL + '/logo192.png';

export const ItemList: React.FC<{}> = () => {
    const [items, setItems] = useState<Item[]>([])
    const fetchItems = () => {
        fetch(server.concat('/items'),
            {
                method: 'GET',
                mode: 'cors',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
            })
            .then(response => response.json())
            .then(data => {
                console.log('GET success:', data);
                setItems(data.items);
            })
            .catch(error => {
                console.error('GET error:', error)
            }
        )
    }

    const fetchImage = (image: string): string => {
        fetch(server.concat('/image/').concat(image),
            {
                method: 'GET',
                mode: 'cors',
                headers: {
                    'Content-Type': 'image/jpeg',
                    'Accept': 'image/jpeg	',
                },
            })
            .then(response => response)
            .then(data => {
                return data.url
            })
            .catch(error => {
                console.error('GET error:', error)
            })
        return server.concat('/image/').concat(image)
    }

    useEffect(() => {
        fetchItems();
    }, []);

    return (
        <div>
            <Listing/>
            {items && items.map((item, index) => {
                return (
                    <div key={index} className='ItemList'>
                        <Link to={"/item/" + item.id}>
                        {/* TODO: Task 1: Replace the placeholder image with the item image */}
                            <img src={fetchImage(item.image)}/>
                        <p>
                            <span>Name: {item.name}</span>
                            <br/>
                            <span>Category: {item.category}</span>
                        </p>
                    </Link>
            </div>
            )
            })}
        </div>
    )
};
