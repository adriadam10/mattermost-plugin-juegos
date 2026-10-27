import React from 'react';

import {apiBase} from '../api.js';

const SUITS = {o: 'oros', c: 'copas', e: 'espadas', b: 'bastos'};

function Card({card, onClick, dim, small}) {
    const src = `${apiBase()}/cards/${card || 'back'}.webp`;
    if (!onClick) {
        return (
            <img
                className={`jg-naipe ${small ? 'jg-naipe-s' : ''} ${dim ? 'jg-dim' : ''}`}
                src={src}
                alt={card || 'carta'}
            />
        );
    }
    return (
        <button
            className='jg-naipe-btn'
            onClick={onClick}
        >
            <img
                className='jg-naipe'
                src={src}
                alt={card}
            />
        </button>
    );
}

// The table from the player's seat: partner opposite, opponents left and right.
export default function Guinote({match, view, send}) {
    const me = view.mine < 0 ? 0 : view.mine;
    const at = (offset) => (me + offset) % 4;
    const played = (seat) => view.trick?.find((p) => p.seat === seat)?.card;
    const over = match.phase === 'fin';
    const myTurn = !over && view.mine >= 0 && view.turn === view.mine;
    const seat = (s, where) => (
        <div className={`jg-seat jg-seat-${where} ${view.turn === s && !over ? 'jg-on-turn' : ''}`}>
            <span>{`${view.seats[s].name}${s === view.dealer ? ' 🂠' : ''}`}</span>
            <span className='jg-muted'>{`${view.seats[s].cards} cartas · pareja ${s % 2 === me % 2 ? 'nuestra' : 'rival'}`}</span>
        </div>
    );
    const ours = me % 2;
    return (
        <div className='jg-col'>
            <div className='jg-row'>
                <span className='jg-tag'>{`Pinta: ${view.trump}`}</span>
                <span className='jg-tag'>{`Juegos ${view.games[ours]}–${view.games[1 - ours]}`}</span>
                {view.vueltas && <span className='jg-tag'>{`De vueltas: ${view.carry[ours]}–${view.carry[1 - ours]} de idas`}</span>}
                {(view.cantes[0] > 0 || view.cantes[1] > 0) && <span className='jg-tag'>{`Cantes ${view.cantes[ours]}–${view.cantes[1 - ours]}`}</span>}
                {view.deck === 0 && !view.muestra && <span className='jg-tag'>{'🔒 Arrastre'}</span>}
            </div>
            <div className='jg-mesa'>
                {seat(at(2), 'top')}
                {seat(at(1), 'right')}
                {seat(at(3), 'left')}
                {view.mine >= 0 ? null : seat(at(0), 'bottom')}
                <div className='jg-center'>
                    {[2, 3, 0, 1].map((o) => (
                        <div
                            key={o}
                            className={`jg-slot jg-slot-${o}`}
                        >
                            {played(at(o)) && <Card card={played(at(o))}/>}
                        </div>
                    ))}
                </div>
                <div className='jg-muestra'>
                    {view.muestra && <Card card={view.muestra}/>}
                    <span className='jg-muted'>{view.deck ? `${view.deck} en el mazo` : 'Sin mazo'}</span>
                </div>
            </div>
            {view.hand?.length > 0 && (
                <div className='jg-hand'>
                    {view.hand.map((c) => {
                        const ok = myTurn && view.legal?.includes(c);
                        return (
                            <Card
                                key={c}
                                card={c}
                                dim={myTurn && !ok}
                                onClick={ok ? () => send('move', {card: c}) : null}
                            />
                        );
                    })}
                </div>
            )}
            {myTurn && <span className='jg-turn'>{view.deck === 0 && !view.muestra ? '👉 Te toca. En el arrastre hay que asistir y montar.' : '👉 Te toca.'}</span>}
            {(view.cante_suits?.length > 0 || view.tute || view.change) && (
                <div className='jg-row'>
                    {view.cante_suits.map((s) => (
                        <button
                            key={s}
                            className='jg-btn'
                            onClick={() => send('move', {cante: s})}
                        >
                            {SUITS[s] === view.trump ? '🎤 Cantar las 40' : `🎤 Cantar las 20 en ${SUITS[s]}`}
                        </button>
                    ))}
                    {view.tute && (
                        <button
                            className='jg-btn'
                            onClick={() => send('move', {tute: true})}
                        >
                            {'👑 ¡Tute!'}
                        </button>
                    )}
                    {view.change && (
                        <button
                            className='jg-btn jg-ghost'
                            onClick={() => send('move', {cambiar: true})}
                        >
                            {'🔄 Cambiar el 7'}
                        </button>
                    )}
                </div>
            )}
            {view.last?.length > 0 && (
                <div className='jg-row'>
                    <span className='jg-muted'>{'Última baza:'}</span>
                    {view.last.map((p) => (
                        <Card
                            key={p.card}
                            card={p.card}
                            small={true}
                        />
                    ))}
                </div>
            )}
            <div className='jg-muted'>
                {(over ? view.history : view.log).map((line, i) => <div key={i}>{line}</div>)}
            </div>
        </div>
    );
}
