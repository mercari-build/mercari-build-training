import { useEffect, useState } from 'react';
import { Item, fetchItems } from '~/api';

const PLACEHOLDER_IMAGE = import.meta.env.VITE_FRONTEND_URL + '/logo192.png';

interface Prop {
  reload: boolean;
  onLoadCompleted: () => void;
  keyword: string;
}

export const ItemList = ({ reload, onLoadCompleted }: Prop) => {
  const [items, setItems] = useState<Item[]>([]);
  useEffect(() => {
    const fetchData = () => {
      fetchItems()
        .then((data) => {
          console.debug('GET success:', data);
          setItems(data.items);
          onLoadCompleted();
        })
        .catch((error) => {
          console.error('GET error:', error);
        });
    };

    if (reload) {
      fetchData();
    }
  }, [reload, onLoadCompleted]);

  const filterItems = (items: Item[], keyword: string) => {
    return items.filter((item) => item.name.toLowerCase().includes(keyword));
  };

  return (
    <div className='ItemListContainer'>
      {filterItems(items, keyword)?.map((item) => {
        console.debug(`Rendering item:`, item); // Debugging log
        // Modified to use the new endpoint that retrieves images by item ID
        const imageUrl = item.id ? `http://localhost:9000/image/id/${item.id}` : PLACEHOLDER_IMAGE;
        
        return (
          <div key={item.id} className="ItemList">
            <img className='item-image'
            src={imageUrl} 
            alt={item.name || 'Item image'} 
            onError={(e) => {
              console.error(`Image failed to load for item ${item.id}, using placeholder.`);
              e.currentTarget.src = PLACEHOLDER_IMAGE;
              }}
            />
            <p>
              <span>Name: {item.name || 'Unknown'}</span>
              <br />
              <span>Category: {item.category || 'Uncategorized'}</span>
            </p>
          </div>
        );
      })}
    </div>
  );
};
