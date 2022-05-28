import {Link, useParams} from "react-router-dom";
import React, {useEffect, useState} from "react";

const server = process.env.API_URL || 'http://127.0.0.1:9000';

export const Item: React.FC<{}> = () => {
    const {itemId} = useParams();
    const [itemName, setItemName] = useState("")
    const [itemCategory, setItemCategory] = useState("")
    const [itemImage, setItemImage] = useState("")
    const fetchItems = () => {
        fetch(server.concat('/items/' + itemId),
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
                setItemName(data.name);
                setItemCategory(data.category);
                setItemImage(data.image);
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
    });
    return (
        <div>
            <h2>{itemName}</h2>
            {/*<div>{itemCategory}</div>*/}
            <img src={fetchImage(itemImage)}/>

            <Link to={"/"}>
                <div>もどる</div>
            </Link>
        </div>
    )
};