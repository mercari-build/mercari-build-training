import { useRef, useState } from 'react';
import { postItem } from '~/api';
import { motion } from 'framer-motion';

interface Prop {
  onListingCompleted: () => void;
}

type FormDataType = {
  name: string;
  category: string;
  image: string | File;
};

export const Listing = ({ onListingCompleted }: Prop) => {
  const initialState = {
    name: '',
    category: '',
    image: '',
  };
  const [values, setValues] = useState<FormDataType>(initialState);
  const [preview, setPreview] = useState<string | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const uploadImageRef = useRef<HTMLInputElement>(null);

  const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values,
      [event.target.name]: event.target.value,
    });
  };
  
  const onFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.files && event.target.files[0]) {
      const file = event.target.files[0];
      setValues({
        ...values,
        [event.target.name]: file,
      });
      
      // Create preview
      const reader = new FileReader();
      reader.onloadend = () => {
        setPreview(reader.result as string);
      };
      reader.readAsDataURL(file);
    }
  };
  
  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setIsSubmitting(true);

    // Validate field before submit
    const REQUIRED_FIELDS = ['name', 'image'];
    const missingFields = Object.entries(values)
      .filter(([key, value]) => !value && REQUIRED_FIELDS.includes(key))
      .map(([key]) => key);

    if (missingFields.length) {
      alert(`Missing fields: ${missingFields.join(', ')}`);
      setIsSubmitting(false);
      return;
    }

    // Submit the form
    postItem({
      name: values.name,
      category: values.category,
      image: values.image,
    })
      .then(() => {
        alert('Item listed successfully');
      })
      .catch((error) => {
        console.error('POST error:', error);
        alert('Failed to list this item');
      })
      .finally(() => {
        onListingCompleted();
        setValues(initialState);
        setPreview(null);
        setIsSubmitting(false);
        if (uploadImageRef.current) {
          uploadImageRef.current.value = '';
        }
      });
  };
  
  return (
    <div className="ListingContainer">
      <h2 className="ListingTitle">Add New Item</h2>
      <motion.div 
        className="Listing"
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <form onSubmit={onSubmit} className="ListingForm">
          <div className="FormFields">
            <div className="InputGroup">
              <label htmlFor="name">Item Name</label>
              <input
                type="text"
                name="name"
                id="name"
                placeholder="Enter item name"
                onChange={onValueChange}
                required
                value={values.name}
                className="FormInput"
              />
            </div>
            
            <div className="InputGroup">
              <label htmlFor="category">Category</label>
              <input
                type="text"
                name="category"
                id="category"
                placeholder="Enter category"
                onChange={onValueChange}
                value={values.category}
                className="FormInput"
              />
            </div>
            
            <div className="InputGroup">
              <label htmlFor="image">Item Image</label>
              <input
                type="file"
                name="image"
                id="image"
                onChange={onFileChange}
                required
                ref={uploadImageRef}
                className="FormInput FileInput"
                accept="image/*"
              />
            </div>
            
            {preview && (
              <div className="ImagePreview">
                <img src={preview} alt="Preview" />
              </div>
            )}
            
            <motion.button 
              type="submit" 
              className="SubmitButton"
              whileHover={{ scale: 1.05 }}
              whileTap={{ scale: 0.95 }}
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Listing...' : 'List this item'}
            </motion.button>
          </div>
        </form>
      </motion.div>
    </div>
  );
};
