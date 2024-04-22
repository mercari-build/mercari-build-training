import React, { useEffect, useState } from 'react';

interface Item {
  id: number;
  name: string;
  category: string;
  image_name: string;  // This field might still exist, but will not be used for images now.
}

const server = process.env.REACT_APP_API_URL || 'http://127.0.0.1:9000';
const placeholderImage = process.env.PUBLIC_URL + '/logo192.png';

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
}

export const ItemList: React.FC<Prop> = (props) => {
  const { reload = true, onLoadCompleted } = props;
  const [items, setItems] = useState<Item[]>([]);
  const fetchItems = () => {
    fetch(`${server}/items`, {
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
      onLoadCompleted && onLoadCompleted();
    })
    .catch((error) => {
      console.error('GET error:', error);
    });
  };

  const getSrcImg = (itemId: number) => {
    return `${server}/image/${itemId}.jpg`;
  };

  useEffect(() => {
    if (reload) {
      fetchItems();
    }
  }, [reload]);

  return (
    <div className="ItemList">
      {items.map((item) => {
        return (
          <div key={item.id} className="Item">
            <img src={getSrcImg(item.id)} alt={item.name} style={{ width: '100px', height: '100px' }} />
            <p>
              <span>Name: {item.name}</span>
              <br />
              <span>Category: {item.category}</span>
            </p>
          </div>
        );
      })}
    </div>
  );
};
