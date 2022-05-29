import React, {useEffect, useState} from "react";
import {QaList} from "./QaList";

export const QaModal = (props: { showFlag: any; setShowQaModal: any; }) => {
    const closeModal = () => {
        props.setShowQaModal(false);
        setIsContract(false);
        setPriceCutModal(false);
    };
    const [showPriceCut, setPriceCutModal] = useState(false); // PriceCutコンポーネントの表示の状態を定義する
    const [isContract, setIsContract] = useState(false); // 値引き交渉成立かどうかのステータス
    const ShowPriceCutModal = () => {
        setPriceCutModal(true);
    };

    const setNegotiation = () => {
        //    todo 金額の値を渡して条件分岐する
        //    成立していたらsetIsContractでvalueをtrue、しなければfalse
        setIsContract(true)
    }
    return (
        <>
            {props.showFlag ? ( // showFlagがtrueだったらModalを表示する
                <div id="overlay" style={{
                    position: "fixed",
                    top: 0,
                    left: 0,
                    width: "100%",
                    height: "100%",
                    backgroundColor: "rgba(0,0,0,0.5)",

                    display: "flex",
                    alignItems: "center",
                    justifyContent: "center",
                }}>
                    <div id="modalContent" style={{
                        background: "white",
                        padding: "10px",
                        borderRadius: "3px",
                    }}>
                        {isContract ? (
                            // 値引き交渉成立の場合
                            <>
                                <div>交渉成立</div>
                                <button onClick={closeModal}>Close</button>
                            </>
                        ) : (
                            <>
                                {showPriceCut ? (
                                    // 値切りフェースの場合
                                    <>
                                        <div>値切り交渉だよ</div>
                                        <button onClick={closeModal}>Close</button>
                                        <button onClick={setNegotiation}>Next</button>
                                    </>
                                ) : (
                                    // QAフェーズの場合
                                    <>
                                        <div><QaList/></div>
                                        <button onClick={closeModal}>Close</button>
                                        <button onClick={ShowPriceCutModal}>Next</button>
                                    </>
                                )}
                            </>
                        )}
                    </div>
                </div>
            ) : (
                <></>// showFlagがfalseの場合はModalは表示しない
            )}
        </>
    )
}