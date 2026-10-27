import React from 'react';

// DuelHead says who plays against whom and whose turn it is, in one-against-one games.
export function DuelHead({match, view, marks}) {
    const over = match.phase === 'fin' || match.phase === 'cancelada';
    let line = '';
    if (!over && match.phase === 'turno') {
        line = view.mine === view.turn ? '👉 Te toca' : `Turno de ${view.players[view.turn]}`;
    }
    return (
        <div className='jg-row'>
            <strong>{`${marks[0]} ${view.players[0]}`}</strong>
            <span className='jg-muted'>{'contra'}</span>
            <strong>{`${marks[1]} ${view.players[1]}`}</strong>
            {line && <span className={view.mine === view.turn ? 'jg-turn' : 'jg-muted'}>{line}</span>}
        </div>
    );
}
