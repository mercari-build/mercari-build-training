import {Outlet} from 'react-router-dom';

// 親コンポーネント
function Items() {
    return (
        <>
            <h2>商品詳細ページ</h2>
            <Outlet/>
        </>
    );
}

export default Items;