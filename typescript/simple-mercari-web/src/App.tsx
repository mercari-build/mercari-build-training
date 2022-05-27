import React from 'react';
import Item from './pages/item';
import {ItemList} from './components/ItemList';
import {BrowserRouter, Route, Routes} from 'react-router-dom';
import './App.css';

function App() {
    return (
        <BrowserRouter>
            <header className='Title'>
                <p>
                    <b>Simple Mercari</b>
                </p>
            </header>
            <Routes>

                <Route path={`/`} element={<ItemList/>}/>
                <Route path={"/about"} element={<About/>}/>
                <Route path={"/item/:index"} element={<Item/>}/>
                <Route path={"/contact"} element={<Contact/>}/>
            </Routes>
        </BrowserRouter>
    );
}

function About() {
    return <h2>About</h2>;
}

function Contact() {
    return <h2>Contact</h2>;
}

export default App;