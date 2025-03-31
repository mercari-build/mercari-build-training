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

    <div className="ItemList">
      {items.map((item) => (
        <div key={item.id} className="card"> {/* ✅ 修正: 商品カードのクラス名を統一 */}
          <img

            src={item.image_name ? `${import.meta.env. VITE_BACKEND_URL}/${item.image_name}` : PLACEHOLDER_IMAGE}

            alt={item.name}
          />
          <div className="item-info">
            <h3>{item.name}</h3>
            <p>{item.category}</p>

          </div>
        </div> // 修正: ここに閉じタグを追加
      ))}
    </div>
  );
}
