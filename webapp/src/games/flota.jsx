import React, {useState} from 'react';

import {DuelHead} from '../match/duel.jsx';

import {randomFleet, shipAt, validPlacement} from './fleet.js';

const SHOT = ['', '·', '🔥', '💥'];

export default function Flota({match, view, send}) {
    if (match.phase === 'colocar') {
        return (
            <Place
                view={view}
                send={send}
            />
        );
    }
    const over = match.phase === 'fin';
    const myTurn = match.phase === 'turno' && view.mine === view.turn;
    const other = view.mine < 0 ? 1 : 1 - view.mine;
    const mineSide = view.mine < 0 ? 0 : view.mine;
    return (
        <div className='jg-col'>
            <DuelHead
                match={match}
                view={view}
                marks={['⚓', '⚓']}
            />
            <div className='jg-boards'>
                <Board
                    title={view.mine < 0 ? `Flota de ${view.players[mineSide]}` : 'Tu flota'}
                    board={view.boards[mineSide]}
                    size={view.size}
                />
                <Board
                    title={view.mine < 0 ? `Flota de ${view.players[other]}` : `Flota de ${view.players[other]}`}
                    board={view.boards[other]}
                    size={view.size}
                    onShoot={myTurn && !over ? (r, c) => send('move', {row: r, column: c}) : null}
                />
            </div>
            <span className='jg-muted'>{'· agua · 🔥 tocado · 💥 hundido. Si tocas, vuelves a disparar.'}</span>
        </div>
    );
}

function Board({title, board, size, onShoot}) {
    const cells = Array.from({length: size * size}, (_, i) => i);
    return (
        <div>
            <strong>{title}</strong>
            <div className='jg-muted'>{`Quedan: ${board.left.length ? board.left.join(', ') : 'ninguno'}`}</div>
            <div
                className='jg-grid'
                style={{gridTemplateColumns: `repeat(${size}, 1fr)`}}
            >
                {cells.map((i) => {
                    const r = Math.floor(i / size);
                    const c = i % size;
                    const shot = board.shots[i];
                    const ship = board.ships && shipAt(board.ships, r, c);
                    return (
                        <button
                            key={i}
                            className={`jg-sea ${ship ? 'jg-ship' : ''} ${board.last === i ? 'jg-last' : ''} ${shot === 1 ? 'jg-water' : ''}`}
                            disabled={!onShoot || shot !== 0}
                            onClick={() => onShoot(r, c)}
                        >
                            {SHOT[shot]}
                        </button>
                    );
                })}
            </div>
        </div>
    );
}

// Place lets a player lay out the fleet: click a cell to put the next ship from
// there, across or down; or at random.
function Place({view, send}) {
    const [ships, setShips] = useState([]);
    const [across, setAcross] = useState(true);
    const [error, setError] = useState('');
    if (view.mine < 0) {
        return <span className='jg-muted'>{'Los jugadores están colocando su flota.'}</span>;
    }
    if (view.placed[view.mine]) {
        return <span className='jg-muted'>{'Flota colocada. Esperando al otro jugador…'}</span>;
    }
    const next = view.fleet[ships.length];
    const put = (r, c) => {
        if (next === undefined) {
            return;
        }
        const ship = Array.from({length: next}, (_, i) => (across ? [r, c + i] : [r + i, c]));
        if (!validPlacement([...ships, ship], view.size)) {
            setError('Ahí no cabe: los barcos no se pueden tocar, ni en diagonal');
            return;
        }
        setError('');
        setShips([...ships, ship]);
    };
    const cells = Array.from({length: view.size * view.size}, (_, i) => i);
    return (
        <div className='jg-col'>
            <span>
                {next === undefined ? '¡Lista! Mándala cuando quieras.' : `Coloca un barco de ${next}: pincha en la casilla donde empieza.`}
            </span>
            <div
                className='jg-grid'
                style={{gridTemplateColumns: `repeat(${view.size}, 1fr)`}}
            >
                {cells.map((i) => {
                    const r = Math.floor(i / view.size);
                    const c = i % view.size;
                    return (
                        <button
                            key={i}
                            className={`jg-sea ${shipAt(ships, r, c) ? 'jg-ship' : ''}`}
                            onClick={() => put(r, c)}
                        />
                    );
                })}
            </div>
            <div className='jg-row'>
                <button
                    className='jg-btn jg-ghost'
                    onClick={() => setAcross(!across)}
                >
                    {across ? '↔ En horizontal' : '↕ En vertical'}
                </button>
                <button
                    className='jg-btn jg-ghost'
                    onClick={() => setShips(randomFleet(view.fleet, view.size))}
                >
                    {'🎲 Al azar'}
                </button>
                <button
                    className='jg-btn jg-ghost'
                    disabled={!ships.length}
                    onClick={() => setShips(ships.slice(0, -1))}
                >
                    {'↶ Quitar el último'}
                </button>
                <button
                    className='jg-btn'
                    disabled={next !== undefined}
                    onClick={() => send('move', {ships})}
                >
                    {'Listo ⚓'}
                </button>
            </div>
            {error && <span className='jg-error'>{error}</span>}
            <span className='jg-muted'>{'Si se acaba el tiempo, se coloca sola al azar.'}</span>
        </div>
    );
}
