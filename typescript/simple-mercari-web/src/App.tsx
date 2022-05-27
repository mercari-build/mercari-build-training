import { useState } from "react";
import "./App.css";
import { ItemList } from "./components/ItemList";
import { Listing } from "./components/Listing";
import { RiRainbowFill } from "react-icons/ri";
import { BrowserRouter as Router, Route, Routes, Link } from "react-router-dom";

function App() {
  // reload ItemList after Listing complete
  const [reload, setReload] = useState(true);
  return (
    <Router>
      <div className="navbar">
        <header className="Title">
          <p className="label">
            <Link to="/items">
              <RiRainbowFill className="icon" />
              <b>Simple Mercari</b>
            </Link>
          </p>
        </header>
        <Listing onListingCompleted={() => setReload(true)} />
      </div>
      <div>
        <ItemList reload={reload} onLoadCompleted={() => setReload(false)} />
      </div>
    </Router>
  );
}

export default App;
