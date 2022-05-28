import { useState } from "react";
import "./App.css";
import { ItemList } from "./components/ItemList";
import { Listing } from "./components/Listing";
import { RiRainbowFill } from "react-icons/ri";
import { BrowserRouter as Router, Route, Routes, Link } from "react-router-dom";
import { useSearchParams } from "react-router-dom";
import Search from "./Items/Search";

const server = process.env.API_URL || "http://localhost:3000";

function App() {
  // reload ItemList after Listing complete
  const [reload, setReload] = useState(true);
  const [query, setQuery] = useState("");

  let url = new URL(`${server}/search?keyword=${query}`);

  const handleSubmit = (e: any) => {
    //go to url
    e.preventDefault();
    window.location.href = `${url}`;
  };


 
  return (
    <Router>
      <div className="navbar">
        <header className="Title">
          <p className="label">
            <Link to="/">
              <RiRainbowFill className="icon" />
              <b>Simple Mercari</b>
            </Link>
          </p>
          <form onSubmit={handleSubmit}>
              <input
                className="searchbar-input"
                type="text"
                value={query}
                placeholder="Search"
                onChange={(e) => setQuery(e.target.value)}
              />
              <button type="submit" className="searchbar-button">
                search
              </button>
            </form>
        </header>
        <Listing onListingCompleted={() => setReload(true)} />
      </div>
      <Routes>
        <Route
          path="/"
          element={
            <ItemList
              reload={reload}
              onLoadCompleted={() => setReload(false)}
            />
          }
        />
        <Route path="/search" element={<Search />} />
      </Routes>
    </Router>
  );
}

export default App;
