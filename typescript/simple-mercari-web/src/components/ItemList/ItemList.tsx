import React, { useEffect, useState } from "react";
import { StringLiteral } from "typescript";

interface Item {
  id: number;
  name: string;
  category: string;
  image: string;
}

interface Category {
  id: number;
  name: string;
}

const server = process.env.API_URL || "http://127.0.0.1:9000";
const placeholderImage = process.env.PUBLIC_URL + "/logo192.png";

interface Prop {
  reload?: boolean;
  onLoadCompleted?: () => void;
}

export const ItemList: React.FC<Prop> = (props) => {
  const { reload = true, onLoadCompleted } = props;
  const [items, setItems] = useState<Item[]>([]);
  const fetchItems = () => {
    fetch(server.concat("/items"), {
      method: "GET",
      mode: "cors",
      headers: {
        "Content-Type": "application/json",
        Accept: "application/json",
      },
    })
      .then((response) => response.json())
      .then((data) => {
        console.log("GET success:", data);
        setItems(data);
        onLoadCompleted && onLoadCompleted();
      })
      .catch((error) => {
        console.error("GET error:", error);
      });
  };

  useEffect(() => {
    if (reload) {
      fetchItems();
    }
  }, [reload]);

  const categories: any = [];
  if (items.length > 0) {
    items.forEach((item) => {
      if (!categories.includes(item.category)) {
        categories.push(item.category);
      }
    });
  }

  const [selectedCategory, setSelectedCategory] = useState<string>("all");

  return (
    <>
      <div className="categories">
        <h1>Categories</h1>
        <ul>
          <li onClick={() => setSelectedCategory("all")}>
            <a>All</a>
          </li>
          {categories.map((category: any) => {
            return (
              <li key={category} onClick={() => setSelectedCategory(category)}>
                <a>{category}</a>
              </li>
            );
          })}
        </ul>
      </div>
      <div className="AllItems">
        {items
          .filter((item: any) => {
            if (selectedCategory === "all") {
              return true;
            } else {
              return item.category === selectedCategory;
            }
          })
          .map((item: any) => {
            return (
              <div className="item">
                <img
                  className="ItemImage"
                  src={server + `/image/${item.image}` || placeholderImage}
                  alt={item.name}
                />
                <div className="item-info">
                  <h2>{item.name}</h2>
                  <p className="tag">{item.category}</p>
                </div>
              </div>
            );
          })}
      </div>
    </>
  );
};
