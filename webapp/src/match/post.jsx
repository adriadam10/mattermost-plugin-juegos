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
    turno: 'Jugando',
    colocar: 'Colocando la flota',
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
            {!over && match.is_creator && match.phase !== 'lobby' && match.info.turns && (
                <div className='jg-row'>
                    <button
                        className='jg-btn jg-danger'
                        onClick={() => send('cancel')}
                    >
                        {'Cancelar partida'}
                    </button>
                </div>
            )}
            {!over && match.is_creator && match.phase !== 'lobby' && !match.info.turns && (
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
    const rival = match.options?.rival;
    const full = info.max_players && players.length >= info.max_players;
    return (
        <div className='jg-col'>
            <span className='jg-muted'>{info.bots ? `${info.description} Caben ${info.max_players}; los huecos los juega el bot.` : `${info.description} Hacen falta ${info.min_players}; caben ${info.max_players}.`}</span>
            <Players players={players}/>
            {info.options?.map((o) => (
                <LobbyOption
                    key={o.key}
                    option={o}
                    value={match.options?.[o.key] ?? ''}
                    editable={match.is_creator}
                    send={send}
                />
            ))}
            <div className='jg-row'>
                {!match.joined && !full && (!rival || match.me === rival) && (
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

// LobbyOption shows an option the creator picks before starting: a choice, or
// whom they challenge.
function LobbyOption({option, value, editable, send}) {
    const [rival, setRival] = useState(value);
    if (!option.choices?.length) {
        if (!editable) {
            return value ? <span>{`⚔️ Reto a @${value}`}</span> : null;
        }
        return (
            <form
                className='jg-row'
                onSubmit={(e) => {
                    e.preventDefault();
                    send('option', {key: option.key, value: rival});
                }}
            >
                <span>{`${option.label}:`}</span>
                <input
                    className='jg-input'
                    value={rival}
                    placeholder='@usuario (vacío: el primero que se apunte)'
                    onChange={(e) => setRival(e.target.value)}
                />
                <button
                    className='jg-btn jg-ghost'
                    type='submit'
                    disabled={rival.replace(/^@/, '') === value}
                >
                    {'Retar'}
                </button>
            </form>
        );
    }
    return (
        <div className='jg-row'>
            <span>{`${option.label}:`}</span>
            {option.choices.map((c) => (
                <button
                    key={c.value}
                    className={`jg-option ${value === c.value ? 'jg-chosen' : ''}`}
                    disabled={!editable}
                    onClick={() => send('option', {key: option.key, value: c.value})}
                >
                    {c.label}
                </button>
            ))}
        </div>
    );
}
