import React from 'react';
import {Item} from './pages/Item';
import NoMatch from './pages/NonMatch';
import {ItemList} from './components/ItemList';
import {ListingUser} from "./components/Listing";
import {BrowserRouter, Route, Routes} from 'react-router-dom';
import './App.css';
import ParentItem from "./pages/ParentItem";

function App() {
    return (
        <div className='App'>
            <BrowserRouter>
                <header className='Title'>
                    <p>
                        <b>Simple Mercari</b>
                    </p>
                </header>
                <Routes>
                    <Route path={`/login`} element={<ListingUser/>}/>
                    <Route path={`/`} element={<ItemList/>}/>
                    <Route path={"/item"} element={<ParentItem/>}>
                        <Route path=":itemId" element={<Item/>}/>
                    </Route>
                    <Route path="*" element={<NoMatch/>}/>
                </Routes>
            </BrowserRouter>
        </div>
    );
}


export default App;