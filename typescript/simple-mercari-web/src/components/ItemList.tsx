import { useEffect, useState } from 'react';
import { Item, fetchItems } from '~/api';

const PLACEHOLDER_IMAGE = import.meta.env.VITE_FRONTEND_URL + '/logo192.png';

interface Prop {
  reload: boolean;
  onLoadCompleted: () => void;
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

  return (
    <div>
      {items?.map((item, index) => {
      console.debug(`Rendering item:`, item); // Debugging log
        const imageUrl = item.id ? `http://localhost:9000/image/${item.id}.jpg` : PLACEHOLDER_IMAGE;
      
        return (
          <div key={item.id || index} className="ItemList">
            {/* TODO: Task 2: Show item images */}
{/*             <img src={PLACEHOLDER_IMAGE} /> */}
               <img src={imageUrl} alt={item.name || 'Item image'} onError={(e) => (e.currentTarget.src = PLACEHOLDER_IMAGE)} />
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
