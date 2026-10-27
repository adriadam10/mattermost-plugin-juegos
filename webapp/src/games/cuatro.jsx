import React from 'react';

import {DuelHead} from '../match/duel.jsx';

const DISCS = ['', '🔴', '🟡'];

export default function Cuatro({match, view, send}) {
    const myTurn = match.phase === 'turno' && view.mine === view.turn;
    const inLine = (r, c) => view.line?.some(([lr, lc]) => lr === r && lc === c);
    const rows = Array.from({length: view.rows}, (_, r) => r);
    const cols = Array.from({length: view.cols}, (_, c) => c);
    return (
        <div className='jg-col'>
            <DuelHead
                match={match}
                view={view}
                marks={['🔴', '🟡']}
            />
            <div
                className='jg-c4'
                style={{gridTemplateColumns: `repeat(${view.cols}, 1fr)`}}
            >
                {rows.map((r) => cols.map((c) => {
                    const v = view.cells[(r * view.cols) + c];
                    const full = view.cells[c] !== 0;
                    return (
                        <button
                            key={`${r}-${c}`}
                            className={`jg-c4-cell ${inLine(r, c) ? 'jg-c4-line' : ''} ${view.last === c && v && r === firstRow(view, c) ? 'jg-c4-last' : ''}`}
                            disabled={!myTurn || full}
                            title={myTurn && !full ? `Columna ${c + 1}` : ''}
                            onClick={() => send('move', {column: c})}
                        >
                            {DISCS[v]}
                        </button>
                    );
                }))}
            </div>
        </div>
    );
}

// The top disc of a column, to mark the last move.
function firstRow(view, c) {
    for (let r = 0; r < view.rows; r++) {
        if (view.cells[(r * view.cols) + c]) {
            return r;
        }
    }
    return -1;
}
