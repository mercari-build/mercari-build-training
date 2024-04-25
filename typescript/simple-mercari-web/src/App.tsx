import { useState } from 'react';
import './App.css';
import { ItemList } from './components/ItemList';
import { Listing } from './components/Listing';

function App() {
  // reload ItemList after Listing complete
  const [reload, setReload] = useState(true);

  return (
    <div className="App">
      <header className="Title">
        <img src={`${process.env.PUBLIC_URL}/Mercari Logo.png`} alt="Mercari Logo" className="CompanyLogo" />
        <p>
          <b>Simple Mercari Web by Zoey</b>
        </p>
      </header>
      <div>
        <Listing onListingCompleted={() => setReload(true)} />
      </div>
      <div>
        <ItemList reload={reload} onLoadCompleted={() => setReload(false)} />
      </div>
    </div>
  );
}

export default App;
