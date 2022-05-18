# STEP5: Implement a simple Mercari webapp as frontend


## 1. Build local environment
Install Node v16 from below.
(16.15.0 LTS is recommended as of May 2022)

https://nodejs.org/en/

If you would like to install multiple versions of Node, use [nvs](https://github.com/jasongin/nvs).

Run `node -v` and confirm that `v16.0.0` or above is displayed.

Move to the following directory and install dependencies with the following command.
```shell
cd typescript/simple-mercari-web
npm ci
```

After launching the web application with the following command, check your web app from your browser at [http://localhost:3000/](http://localhost:3000/).
```shell
npm start
```

Run the backend servers in Python/Go as described in Step3.
This simple web application allows you to do two things
- Add a new item (Listing)
- View the list of itemas (ItemList)
  
These functionalities are carved out as components called `src/components/Listing` and `src/components/ItemList`, and called from the main `App.tsx`.

:pushpin: Sample code is in React but the knowledge of React is not necessary.

### (Optional) Task 1: Add a new item
Use the listing form to add a new item. In this screen, you can input name, category and a image for a new item.

If your API from STEP3 only accepts name and category, modify `typescript/simple-mercari-web/src/components/Listing/Listing.tsx` and delete the image field.


### (Optional) Task 2. Show item images 

In this screen, item images are all rendered as Build@Mercari logo. Specify the item image as `http://localhost:9000/image/<item_id>.jpg` and see if they can be displayed on the web app.


### (Optional) Task 3. Change the styling with HTML and CSS
These two components are styled by CSS. To see what types of changes can be made, try modifying `ItemList` component CSS. These are specifed in `App.css` and they are applied by `className` attribute (e.g. `<div className='Listing'></div>`).
```css
.Listing {
  ...
}
.ItemList {
  ...
}
```
Try editing the HTML tags returned in each component and see how the UI changes.


### (Optional) Task 4. Change the UI for ItemList

Current `ItemList` shows one column of items sequentially. Use the following reference to change this style into a grid system where multiple items are displayed in the same row.


**:book: References**

- [HTML basics](https://developer.mozilla.org/en-US/docs/Learn/Getting_started_with_the_web/HTML_basics)


- [CSS basics](https://developer.mozilla.org/en-US/docs/Learn/Getting_started_with_the_web/CSS_basics)


- [Basic Concepts of grid layout](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Grid_Layout/Basic_Concepts_of_Grid_Layout)

---

### Next

[STEP6: Run frontend and API using docker-compose](step6.en.md)