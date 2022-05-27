import React, { useEffect, useState } from "react";
import { StringLiteral } from "typescript";
import { AiOutlineSearch } from "react-icons/ai";
import {useSearchParams} from "react-router-dom";


interface Item {
  id: number;
  en_name: string;
  ja_name: string;
  category: string;
  image: string;
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
  // const [search, setSearch] = useState("");
  // const [searchItems, setSearchItems] = useState<Item[]>([]);
  // const fetchSearchItems = () => {
  //   fetch(server.concat("/items?search=").concat(search), {
  //     method: "GET",
  //     mode: "cors",
  //     headers: {
  //       "Content-Type": "application/json",
  //       Accept: "application/json",
  //     },
  //   })
  //     .then((response) => response.json())
  //     .then((data) => {
  //       console.log("GET success:", data);
  //       setSearchItems(data);
  //       onLoadCompleted && onLoadCompleted();
  //     })
  //     .catch((error) => {
  //       console.error("GET error:", error);
  //     });
  // };

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
  const [selectedLanguage, setSelectedLanguage] = useState<string>("en");
  const [searchText, setSearchText] = useState<string>("");

  const handleCategoryChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedCategory(e.target.value);
  };

  const handleLanguageChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    setSelectedLanguage(e.target.value);
  };

  const handleSearchTextChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearchText(e.target.value);
  };

  const filteredItems = items
    .filter((item) => {
      if (selectedCategory === "all") {
        return true;
      } else {
        return item.category === selectedCategory;
      }
    })
    .filter((item) => {
      if (searchText === "") {
        return true;
      } else {
        return (
          item.en_name.includes(searchText) || item.ja_name.includes(searchText)
        );
      }
    });

  return (
    <div className="item-list">
      <div className="item-list-header">
        <div className="item-list-search">
          <input
            className="searchbar-input"
            type="text"
            placeholder="Search"
            value={searchText}
            onChange={handleSearchTextChange}
          />
          <AiOutlineSearch className="searchbar-button" />
        </div>
        <div className="item-list-header-left">
          <select
            className="select"
            value={selectedCategory}
            onChange={handleCategoryChange}
          >
            <option value="all">All</option>
            {categories.map((category: any) => {
              return (
                <option key={category} value={category}>
                  {category}
                </option>
              );
            })}
          </select>
        </div>
        <div className="item-list-header-right">
          <select
            className="select"
            value={selectedLanguage}
            onChange={handleLanguageChange}
          >
            <option value="en">English</option>
            <option value="ja">Japanese</option>
          </select>
        </div>
      </div>
      <div className="AllItems">
        {filteredItems.map((item) => {
          return (
            <div className="item-list-item" key={item.id}>
              <div className="item-list-item-image">
                <img
                  className="ItemImage"
                  src={server + `/image/${item.image}` || placeholderImage}
                  alt={item.en_name}
                />
              </div>
              <div className="item-list-item-name">
                {selectedLanguage === "en" ? item.en_name : item.ja_name}
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};
