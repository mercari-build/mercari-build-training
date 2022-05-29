import {Link, Outlet} from 'react-router-dom';
import React from "react";

// 親コンポーネント
function qaTest() {
    console.log("QAをここに作る")
}
function Items() {
    return (
        <>
            <div style={{
                position: 'relative',
                padding: '1rem 1rem 60px',
            }}>
                <h2>商品詳細ページ</h2>
                <Outlet/>
                <div style={{
                    position: 'absolute',
                }}>
                    <div id="qa" style={{
                        padding: '30px',
                        width: '80px',
                        textAlign: 'center',
                        background: 'skyblue',
                        position: 'fixed',
                        bottom: 55,
                        right: 50,
                        fontSize: 30,
                    }} onClick={qaTest}>QA
                    </div>
                </div>
            </div>
        </>
    );
}

export default Items;