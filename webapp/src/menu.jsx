import React, {useEffect, useState} from 'react';

import {createMatch, fetchGames, fetchRanking} from './api.js';
import {Overlay} from './match/common.jsx';

export default function Menu({channelId, onClose, onCanvas}) {
    const [info, setInfo] = useState(null);
    const [ranking, setRanking] = useState(null);
    const [mode, setMode] = useState('directo');
    const [error, setError] = useState('');
    const [busy, setBusy] = useState(false);

    useEffect(() => {
        fetchGames().then(setInfo, () => setError('No se han podido cargar los juegos'));
        fetchRanking().then(setRanking, () => {});
    }, []);

    const start = async (game) => {
        setBusy(true);
        setError('');
        try {
            const res = await createMatch(game, mode, channelId);
            if (res.refused) {
                setError(res.refused);
                return;
            }
            onClose();
        } catch (err) {
            setError(game === 'pie' ? 'No se ha podido crear la partida. ¿Hay clave de Giphy?' : 'No se ha podido crear la partida');
        } finally {
            setBusy(false);
        }
    };

    return (
        <Overlay onClose={onClose}>
            <h2>{'🎲 Juegos'}</h2>
            <div className='jg-row' style={{marginTop: 0, marginBottom: 12}}>
                <button
                    className={`jg-option ${mode === 'directo' ? 'jg-chosen' : ''}`}
                    onClick={() => setMode('directo')}
                >
                    {'⚡ En directo'}
                    {info && <div className='jg-muted'>{`Fases de ~${info.live_seconds} s`}</div>}
                </button>
                <button
                    className={`jg-option ${mode === 'async' ? 'jg-chosen' : ''}`}
                    onClick={() => setMode('async')}
                >
                    {'🐢 Asíncrona'}
                    {info && <div className='jg-muted'>{`Fases de ~${info.async_hours} h, entre tarea y tarea`}</div>}
                </button>
            </div>
            <div className='jg-games'>
                {info?.games.map((g) => (
                    <button
                        key={g.id}
                        className='jg-option jg-game'
                        disabled={busy || (g.id === 'pie' && !info.giphy_enabled)}
                        onClick={() => start(g.id)}
                    >
                        <span className='jg-emoji'>{g.emoji}</span>
                        <strong>{g.name}</strong>
                        <span className='jg-muted'>{g.description}</span>
                        <span className='jg-muted'>{players(g)}</span>
                    </button>
                ))}
                <button
                    className='jg-option jg-game'
                    onClick={onCanvas}
                >
                    <span className='jg-emoji'>{'👾'}</span>
                    <strong>{'Lienzo de temporada'}</strong>
                    <span className='jg-muted'>{'Un lienzo de píxeles para todo el servidor. Pon un píxel de vez en cuando; a final de mes se publica.'}</span>
                </button>
            </div>
            {error && <p className='jg-error'>{error}</p>}
            {ranking && <Ranking ranking={ranking}/>}
        </Overlay>
    );
}

function players(g) {
    if (!g.lobby) {
        return `Juega quien quiera, desde ${g.min_players}`;
    }
    return `De ${g.min_players} a ${g.max_players} jugadores`;
}

function Ranking({ranking}) {
    if (!ranking.rows.length) {
        return <p className='jg-muted'>{`Nadie ha puntuado aún en la temporada ${ranking.season}.`}</p>;
    }
    return (
        <div style={{marginTop: 16}}>
            <strong>{`🏆 Temporada ${ranking.season}`}</strong>
            <table className='jg-table'>
                <tbody>
                    {ranking.rows.slice(0, 10).map((r, i) => (
                        <tr key={r.username}>
                            <td>{['🥇', '🥈', '🥉'][i] ?? i + 1}</td>
                            <td>{`@${r.username}`}</td>
                            <td>{`${r.points} pts`}</td>
                            <td className='jg-muted'>{`${r.played} partidas · ${r.wins} victorias`}</td>
                        </tr>
                    ))}
                </tbody>
            </table>
        </div>
    );
}
