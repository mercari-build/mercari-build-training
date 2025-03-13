import { useState } from 'react';
import './App.css';
import { ItemList } from '~/components/ItemList';
import { Listing } from '~/components/Listing';
import { motion } from 'framer-motion';

function App() {
  // reload ItemList after Listing complete
  const [reload, setReload] = useState(true);
  return (
    <div className="App">
      <motion.header 
        className="Title"
        initial={{ opacity: 0, y: -50 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <h1>Mercari</h1>
      </motion.header>
      
      <main className="AppContent">
        <Listing onListingCompleted={() => setReload(true)} />
        <ItemList reload={reload} onLoadCompleted={() => setReload(false)} />
      </main>
      
      <footer className="Footer">
        <p>© 2023 Simple Mercari</p>
      </footer>
    </div>
  );
}

export default App;
