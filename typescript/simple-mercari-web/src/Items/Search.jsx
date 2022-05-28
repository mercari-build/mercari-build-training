import React from 'react';
import { useEffect, useState } from "react";
import { useSearchParams } from "react-router-dom";
import { AiOutlineSearch } from 'react-icons/ai';


const server = process.env.API_URL || "http://127.0.0.1:9000/search";
const placeholderImage = process.env.PUBLIC_URL + "/logo192.png";

export default function Search() {
  const [searchParams, setSearchParams] = useSearchParams();
  const searchKeyword = searchParams.get("keyword");
  
    const [items, setItems] = useState([]);
    const fetchItems = async () => {
        const response = await fetch(`${server}?keyword=${searchKeyword}`);
        const json = await response.json();
        setItems(json.items);
        console.log(json);
    };
    useEffect(() => {
        fetchItems();
    }, [searchKeyword]);



    const categories = [];
    if (items.length > 0) {
      items.forEach((item) => {
        if (!categories.includes(item.category)) {
          categories.push(item.category);
        }
      });
    }

  
  const [selectedCategory, setSelectedCategory] = useState("all");
  const [selectedLanguage, setSelectedLanguage] = useState("en");
  const [searchText, setSearchText] = useState("");

  const handleCategoryChange = (e) => {
    setSelectedCategory(e.target.value);
  };

  const handleLanguageChange = (e) => {
    setSelectedLanguage(e.target.value);
  };

  const handleSearchTextChange = (e) => {
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
            {categories.map((category) => {
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
                  src={process.env.API_URL || "http://127.0.0.1:9000" + `/image/${item.image}` || placeholderImage}
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

