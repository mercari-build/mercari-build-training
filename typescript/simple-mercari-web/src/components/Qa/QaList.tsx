import React, {useEffect, useState} from 'react';
interface Qa {
    id: number;
    item_id: number;
    question: string;
    answer: string;
    qa_type_id: number;
}

const server = process.env.API_URL || 'http://127.0.0.1:9000';

export const QaList: React.FC<{}> = () => {
    const [qas, setQas] = useState<Qa[]>([])
    const fetchQas = () => {
        // todo item_id取得
        fetch(server.concat('/qas/333'),
            {
                method: 'GET',
                mode: 'cors',
                headers: {
                    'Content-Type': 'application/json',
                    'Accept': 'application/json'
                },
            })
            .then(response => response.json())
            .then(data => {
                console.log('GET success:', data);
                setQas(data.qas);
            })
            .catch(error => {
                console.error('GET error:', error)
            })
    }

    useEffect(() => {
        fetchQas();
    }, []);

    return (
        <div>
            {qas && qas.map((qa) => {
                return (
                    <div key={qa.id} className='QaList'>
                        <p>
                            <span>Question: {qa.question}</span>
                            <br/>
                            <span>Answer: {qa.answer}</span>
                        </p>
                    </div>
                )
            })}
        </div>
    )
}