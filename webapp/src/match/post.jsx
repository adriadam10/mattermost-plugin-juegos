import React, {useCallback, useEffect, useState} from 'react';

import {act, fetchMatch} from '../api.js';
import {on} from '../bus.js';
import {GAMES} from '../games/index.js';

import {Countdown, Players, Scores} from './common.jsx';

// Phase names shared by the games; an unknown one is shown as is.
const PHASES = {
    lobby: 'Esperando jugadores',
    elegir: 'Eligiendo',
    responder: 'Respondiendo',
    escribir: 'Escribiendo',
    dibujar: 'Dibujando',
    describir: 'Describiendo',
    votar: 'Votación',
    fin: 'Terminada',
    cancelada: 'Cancelada',
};

// MatchPost draws a match in its channel post and keeps it in step with the server.
export default function MatchPost({post}) {
    const id = post.props?.match_id;
    const [match, setMatch] = useState(null);
    const [offset, setOffset] = useState(0);
    const [error, setError] = useState('');
    const [gone, setGone] = useState(false);

    const load = useCallback(() => {
        fetchMatch(id).then((m) => {
            setMatch(m);
            setOffset(m.now - Date.now());
        }, () => setGone(true));
    }, [id]);

    useEffect(() => {
        if (!id) {
            return undefined;
        }
        load();
        const offMatch = on(`match:${id}`, load);
        const offReconnect = on('reconnect', load);
        return () => {
            offMatch();
            offReconnect();
        };
    }, [id, load]);

    // A phase that ends while nobody acts is moved on by the server within seconds.
    useEffect(() => {
        if (!match?.deadline) {
            return undefined;
        }
        const wait = Math.max(match.deadline - (Date.now() + offset), 0) + 6000;
        const t = setTimeout(load, wait);
        return () => clearTimeout(t);
    }, [match?.deadline, offset, load]);

    const send = useCallback(async (action, move) => {
        setError('');
        try {
            const res = await act(id, action, move);
            if (res.refused) {
                setError(res.refused);
                return false;
            }
            setMatch(res);
            return true;
        } catch (err) {
            setError('No se ha podido enviar. Prueba otra vez.');
            return false;
        }
    }, [id]);

    if (gone) {
        return <div className='jg-card jg-muted'>{'Esta partida ya no está disponible; los resultados están en el hilo.'}</div>;
    }
    if (!match) {
        return <div className='jg-card jg-muted'>{'Cargando partida…'}</div>;
    }
    const Game = GAMES[match.info.id];
    const over = match.phase === 'fin' || match.phase === 'cancelada';
    return (
        <div className='jg-card'>
            <div className='jg-head'>
                <span className='jg-title'>{`${match.info.emoji} ${match.info.name}`}</span>
                <span className='jg-tag'>{match.mode === 'async' ? '🐢 asíncrona' : '⚡ en directo'}</span>
                <span className='jg-tag'>{PHASES[match.phase] ?? match.phase}</span>
                <Countdown
                    deadline={match.deadline}
                    offset={offset}
                />
            </div>
            {match.phase === 'lobby' && (
                <Lobby
                    match={match}
                    send={send}
                />
            )}
            {match.game && Game && (
                <Game
                    match={match}
                    view={match.game}
                    send={send}
                />
            )}
            {match.scores && <Scores scores={match.scores}/>}
            {error && <p className='jg-error'>{error}</p>}
            {!over && match.is_creator && match.phase !== 'lobby' && (
                <div className='jg-row'>
                    <button
                        className='jg-btn jg-ghost'
                        onClick={() => send('skip')}
                    >
                        {'Pasar de fase ⏭'}
                    </button>
                    <button
                        className='jg-btn jg-danger'
                        onClick={() => send('cancel')}
                    >
                        {'Cancelar partida'}
                    </button>
                </div>
            )}
        </div>
    );
}

function Lobby({match, send}) {
    const {info, players} = match;
    const enough = players.length >= info.min_players;
    return (
        <div className='jg-col'>
            <span className='jg-muted'>{`${info.description} Hacen falta ${info.min_players}; caben ${info.max_players}.`}</span>
            <Players players={players}/>
            <div className='jg-row'>
                {!match.joined && (
                    <button
                        className='jg-btn'
                        onClick={() => send('join')}
                    >
                        {'Me apunto 🙋'}
                    </button>
                )}
                {match.joined && !match.is_creator && (
                    <button
                        className='jg-btn jg-ghost'
                        onClick={() => send('leave')}
                    >
                        {'Me borro'}
                    </button>
                )}
                {match.is_creator && (
                    <button
                        className='jg-btn'
                        disabled={!enough}
                        onClick={() => send('start')}
                    >
                        {enough ? 'Empezar ya ▶' : `Faltan ${info.min_players - players.length}`}
                    </button>
                )}
                {match.is_creator && (
                    <button
                        className='jg-btn jg-danger'
                        onClick={() => send('cancel')}
                    >
                        {'Cancelar'}
                    </button>
                )}
            </div>
            <span className='jg-muted'>{'Si se llega al mínimo, empieza sola al acabar la cuenta atrás.'}</span>
        </div>
    );
}
