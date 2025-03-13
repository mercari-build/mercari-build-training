import { useEffect, useState } from 'react';
import { Item, fetchItems } from '~/api';
import { motion } from 'framer-motion';

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
    <div className="ItemListContainer">
      <h2 className="ItemListTitle">Available Items</h2>
      <div className="ItemListGrid">
        {items?.map((item, index) => {
          return (
            <motion.div
              key={item.id}
              className="ItemCard"
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3, delay: index * 0.1 }}
              whileHover={{ scale: 1.05 }}
            >
              <div className="ItemImageContainer">
                <img 
                  src={`http://localhost:9000/images/${item.id}.jpg`} 
                  onError={(e) => {
                    e.currentTarget.src = import.meta.env.VITE_FRONTEND_URL + '/logo192.png';
                  }}
                  alt={item.name}
                  className="ItemImage"
                />
              </div>
              <div className="ItemDetails">
                <h3 className="ItemName">{item.name}</h3>
                <span className="ItemCategory">{item.category}</span>
              </div>
            </motion.div>
          );
        })}
      </div>
    </div>
  );
};
