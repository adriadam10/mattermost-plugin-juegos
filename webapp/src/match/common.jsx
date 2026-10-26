import React, {useEffect, useState} from 'react';

import {remaining} from '../time.js';

export function Overlay({onClose, children}) {
    useEffect(() => {
        const onKey = (e) => e.key === 'Escape' && onClose();
        window.addEventListener('keydown', onKey);
        return () => window.removeEventListener('keydown', onKey);
    }, [onClose]);
    return (
        <div
            className='jg-overlay'
            onMouseDown={(e) => e.target === e.currentTarget && onClose()}
        >
            <div className='jg-modal'>{children}</div>
        </div>
    );
}

// Countdown to the phase deadline, corrected by the server clock offset.
export function Countdown({deadline, offset}) {
    const [now, setNow] = useState(Date.now());
    useEffect(() => {
        const id = setInterval(() => setNow(Date.now()), 1000);
        return () => clearInterval(id);
    }, []);
    if (!deadline) {
        return null;
    }
    const left = deadline - (now + offset);
    return <span className={`jg-clock ${left < 15000 ? 'jg-urgent' : ''}`}>{`⏱ ${remaining(left)}`}</span>;
}

export function Players({players}) {
    return (
        <div className='jg-players'>
            {players.map((p) => (
                <span
                    key={p.id}
                    className='jg-player'
                >
                    {`@${p.username}`}
                </span>
            ))}
        </div>
    );
}

export function Scores({scores}) {
    return (
        <table className='jg-table'>
            <tbody>
                {scores.map((s, i) => (
                    <tr key={s.username}>
                        <td>{['🥇', '🥈', '🥉'][i] ?? i + 1}</td>
                        <td>{`@${s.username}`}</td>
                        <td>{s.score}</td>
                        <td className='jg-muted'>{`+${s.awarded} ranking`}</td>
                    </tr>
                ))}
            </tbody>
        </table>
    );
}

// TextMove is a text box that sends its value as a move, for answers and captions.
export function TextMove({initial = '', placeholder, max = 140, label = 'Enviar', onSend}) {
    const [text, setText] = useState(initial);
    useEffect(() => setText(initial), [initial]);
    const send = (e) => {
        e.preventDefault();
        if (text.trim()) {
            onSend(text.trim());
        }
    };
    return (
        <form
            className='jg-row'
            onSubmit={send}
        >
            <input
                className='jg-input'
                value={text}
                maxLength={max}
                placeholder={placeholder}
                onChange={(e) => setText(e.target.value)}
            />
            <button
                className='jg-btn'
                type='submit'
                disabled={!text.trim() || text.trim() === initial}
            >
                {initial ? 'Cambiar' : label}
            </button>
        </form>
    );
}
