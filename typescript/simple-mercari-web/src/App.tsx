import React from 'react';
import {Item} from './pages/Item';
import Items from './pages/Items';
import NoMatch from './pages/NonMatch';
import {ItemList} from './components/ItemList';
import {BrowserRouter, Route, Routes} from 'react-router-dom';
import './App.css';

function App() {
    return (
        <div className = 'App'>
            <BrowserRouter>
                <header className='Title'>
                    <p>
                        <b>Simple Mercari</b>
                    </p>
                </header>
                <Routes>

                    <Route path={`/`} element={<ItemList/>}/>
                    <Route path={"/item"} element={<Items />}>
                        <Route path=":itemId" element={<Item />} />
                    </Route>
                    <Route path="*" element={<NoMatch />} />
                </Routes>
            </BrowserRouter>
        </div>
    );
}


export default App;