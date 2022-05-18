import { useState } from 'react';
import './App.css';
import { ItemList } from './components/ItemList';
import { Listing } from './components/Listing';

function App() {
  // reload ItemList after Listing complete
  const [reload, setReload] = useState(true);
  return (
    <div>
      <header className='Title header-part'>
        <p>
          <b>Simple Mercari</b>
        </p>
        <div>
          <Listing onListingCompleted={() => setReload(true)} />
        </div>
      </header>
      <div >
        <ItemList reload={reload} onLoadCompleted={() => setReload(false)} />
      </div>
    </div>
  )
}

export default App;
