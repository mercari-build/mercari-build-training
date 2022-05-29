import {Link, Outlet} from 'react-router-dom';
import React, {useEffect, useState} from "react";
import {QaModal} from "../components/Qa/QaModal";

// Itemの親コンポーネント
function qaTest() {
    console.log("QAをここに作る")
}
function ParentItem() {
    const [showQaModal, setShowQaModal] = useState(false); // Modalコンポーネントの表示の状態を定義する
    const showModal = () => {
        setShowQaModal(true);
    };
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
                    <div style={{
                        // padding: '30px',
                        // width: '80px',
                        textAlign: 'center',
                        position: 'fixed',
                        bottom: 150,
                        right: 50,
                        fontSize: 14}}>よくある質問はこちら</div>
                    <div id="qa" style={{
                        padding: '30px',
                        width: '80px',
                        textAlign: 'center',
                        background: 'skyblue',
                        position: 'fixed',
                        bottom: 55,
                        right: 50,
                        fontSize: 30,
                    }} onClick={showModal}>QA
                    </div>
                </div>
                <QaModal showFlag={showQaModal} setShowQaModal={setShowQaModal} />
                {/*<QaModal showFlag={showQaModal} setShowQaModal={setShowQaModal} content="親から渡された値です。"/>*/}
            </div>
        </>
    );
}

export default ParentItem;