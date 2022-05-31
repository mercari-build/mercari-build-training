import React, { useEffect, useState } from 'react';

interface Item {
  id: number;
  name: string;
  category: string;
  image_filename: string;
};

const server = process.env.API_URL || 'http://127.0.0.1:9000';

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
}

export const ItemList: React.FC<Prop> = (props) => {
  const { reload = true, onLoadCompleted } = props;
  const [items, setItems] = useState<Item[]>([]);
  // check if we need empty state (i.e., when nothing is listed)
  const [emptyState, setEmpty] = useState(true);
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
      .then(function (response) {
        if (response.ok) {
          setEmpty(false);
          response.json()
            .then(data => {
              console.log('GET success:', data);
              setItems(data.items);
              onLoadCompleted && onLoadCompleted();
            })
            .catch(error => {
              console.error('GET error:', error)
            })
        } else {
          setEmpty(true);
        }
      })
  }


  useEffect(() => {
    if (reload) {
      fetchItems();
    }
  }, [reload]);

  // if (emptyState) {
  //   return (
  //     <div className="empty-state">
  //       <img id="empty-image" src="emptystate.svg" alt="Empty State" />
  //       <h3>No Listed Item</h3>
  //       <p>Newly listed items will appear here.</p>
  //     </div>
  //   )
  // }
  // else {
  return (
    <div className='GridListing'>
      {items.map((item) => {
        return (
          <div key={item.id} className='ItemList'>
            {/* TODO: Task 1: Replace the placeholder image with the item image */}
            <img src={`${server}/image/${item.image_filename}`} alt="item-image" className='ListedImage' />
            <p>
              <span id="item-category"> {item.category}</span>
              <span id="item-name">{item.name}</span>
              <button>
                <img id="delete-btn" src="delete-icon.png" alt="delete icon"></img>
              </button>
            </p>
          </div>
        )
      })}
    </div>
  )
  // }
};
