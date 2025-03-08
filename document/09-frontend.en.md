# STEP9: Implement a simple Mercari webapp as frontend

## 1. Build local environment

Install Node v22 from below.
(v22.13.1 LTS is recommended as of Feb 2025)

https://nodejs.org/en/

If you would like to install multiple versions of Node, use [nvs](https://github.com/jasongin/nvs).

Run `node -v` and confirm that `v22.0.0` or above is displayed.

Move to the following directory and install dependencies with the following command.

```shell
cd typescript/simple-mercari-web
npm ci
```

After launching the web application with the following command, check your web app from your browser at [http://localhost:3000/](http://localhost:3000/).

```shell
npm start
```

Run the backend servers in Python/Go as described in STEP4.
This simple web application allows you to do two things

- Add a new item (Listing)
- View the list of items (ItemList)

These functionalities are carved out as components called `src/components/Listing.tsx` and `src/components/ItemList.tsx`, and called from the main `App.tsx`.

:pushpin: Sample code is in React but the knowledge of React is not necessary.

## (Optional) Task 1: Add a new item

Use the listing form to add a new item. In this screen, you can input name, category and an image for a new item.

If your API from STEP4 only accepts name and category, modify `typescript/simple-mercari-web/src/components/Listing.tsx` and delete the image field.

## (Optional) Task 2. Show item images

In this screen, item images are all rendered as Build@Mercari logo. Specify the item image as `http://localhost:9000/images/<item_id>.jpg` and see if they can be displayed on the web app.

## (Optional) Task 3. Change the styling with HTML and CSS

These two components are styled by CSS. To see what types of changes can be made, try modifying `ItemList` component CSS. These are specified in `App.css` and they are applied by `className` attribute (e.g. `<div className='Listing'></div>`).

```css
.Listing {
  ...;
}
.ItemList {
  ...;
}
```

Try editing the HTML tags returned in each component and see how the UI changes.

## (Optional) Task 4. Change the UI for ItemList

Current `ItemList` shows one column of items sequentially. Use the following reference to change this style into a grid system where multiple items are displayed in the same row.

**:book: References**

- [HTML basics](https://developer.mozilla.org/en-US/docs/Learn/Getting_started_with_the_web/HTML_basics)

- [CSS basics](https://developer.mozilla.org/en-US/docs/Learn/Getting_started_with_the_web/CSS_basics)

- [Basic Concepts of grid layout](https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Grid_Layout/Basic_Concepts_of_Grid_Layout)

---

## Tips

### Debugging

Debugging is the process of checking the operation of a program, identifying problems, and fixing them. In web front-end development, debugging can be performed using the following methods:

By inserting `console.debug()` at the points in the code where you want to check the operation, you can verify the values and states at runtime. For example, in `ItemList.tsx`:

```typescript
export const ItemList = (props: Prop) => {
  ...
  useEffect(() => {
    const fetchData = () => {
      fetchItems()
        .then((data) => {
          console.debug('GET success:', data); // Check the contents of the data retrieved from the API here
          ...
        })
        .catch((error) => {
          console.error('GET error:', error);
        });
    };
  ...
```

This debugging information can be checked using the browser's developer tools (**Chrome DevTools**). Chrome DevTools can be opened in any of the following ways:

- Keyboard shortcuts:
  - Windows/Linux: `Ctrl + Shift + I`
  - macOS: `Cmd + Option + I`
- Right-click on the browser and select "Inspect"
- From the menu, select "More tools" > "Developer tools"

The information output by `console.debug()` will be displayed in the "Console" tab of the developer tools.

For detailed instructions on how to use the developer tools, refer to the [Chrome DevTools documentation](https://developer.chrome.com/docs/devtools/open?hl=en).

### Build Production-Ready App by using Framework

This material aims to provide a basic understanding of React, so it does not use any specific frameworks. However, the React development team recommends using the following frameworks when developing actual production-level web services ([Creating a React App](https://react.dev/learn/creating-a-react-app)):

- [Next.js (App Router)](https://nextjs.org/docs)
- [React Router (v7)](https://reactrouter.com/start/framework/installation)

:warning: As a point of caution, **it has been [officially announced on February 14, 2025](https://react.dev/blog/2025/02/14/sunsetting-create-react-app) that [`create-react-app`](https://github.com/facebook/create-react-app), which is introduced in many React materials, will be deprecated**. For new projects, it is strongly recommended to consider methods other than using `create-react-app`.

In the future, when you are responsible for developing services intended for long-term use by users, consider using these frameworks. When creating a new service, it is important to understand the characteristics of each framework and the requirements of the service you want to create, and select the appropriate framework.

When selecting a framework, it is good to consider the following perspectives:

- Scale and complexity of the service
- Performance requirements
- SEO requirements
- Team's technology stack
- Deployment environment

## Next

[STEP10: Run frontend and API using docker-compose](./10-docker-compose.en.md)
