import React, { useEffect, useMemo, useState } from 'react';
import './App.css';
import { ItemList } from './components/ItemList';
import { Listing } from './components/Listing';

function App() { 
  return (
    <div>
      <header className='Title'>
        <p>
          <b>Simple Mercari</b>
        </p>
      </header>
      <div>
        <Listing/>
      </div>
      <div>
        <ItemList/>
      </div>
    </div>
  )
}

export default App;