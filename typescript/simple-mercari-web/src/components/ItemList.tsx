import { useEffect, useState } from 'react';
import { Item, fetchItems } from '../api';

const placeholderImage = import.meta.env.VITE_FRONTEND_URL + '/logo192.png';

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
}

export const ItemList: React.FC<Prop> = (props) => {
  const { reload = true, onLoadCompleted } = props;
  const [items, setItems] = useState<Item[]>([]);
  useEffect(() => {
    const fetchData = () => {
      fetchItems()
        .then((data) => {
          console.log('GET success:', data);
          setItems(data.items);
          if (onLoadCompleted) {
            onLoadCompleted();
          }
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
      {items.map((item) => {
        return (
          <div key={item.id} className="ItemList">
            {/* TODO: Task 1: Replace the placeholder image with the item image */}
            <img src={placeholderImage} />
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
