import React, { useEffect, useState } from 'react';

interface Item {
  id: number;
  name: string;
  category: string;
  image_filename: string;
};

const server = process.env.API_URL || 'http://127.0.0.1:9000';
const placeholderImage = process.env.PUBLIC_URL + '/logo192.png';

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
}

export const ItemList: React.FC<Prop> = (props) => {
  const { reload = true, onLoadCompleted } = props;
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
        onLoadCompleted && onLoadCompleted();
      })
      .catch(error => {
        console.error('GET error:', error)
      })
  }

  const fetchImage = (item: Item)=> {
    // TODO: debug: item.image_filename が undefinedになる
    fetch(server.concat('/image/'+item.image_filename),
    {
      method: 'GET',
      mode: 'cors',
      headers : {
        'Content-Type': 'image/jpg',
        'Accept': 'image/jpg	',
      },
    })
      .then(response => response)
      .then(data => {
        return data.url
      })
      .catch(error => {
        console.error('GET error:',error) 
      })

      return server+'/image/'+item.image_filename
    
  }

  useEffect(() => {
    if (reload) {
      fetchItems();
    }
  }, [reload]);

  return (
    <div className='grid'>
      {items.map((item) => {
        return (
          <div key={item.id} className='ItemList'>
            <div className='item'>
            {/* TODO: Task 1: Replace the placeholder image with the item image */}
            <img src={fetchImage(item)|| placeholderImage} />
            {/* <img src={placeholderImage} /> */}
            <p>
              <span>Name: {item.name}</span>
              <br />
              <span>Category: {item.category}</span>
            </p>
            </div>
          </div>
        )
      })}
    </div>
  )
};
