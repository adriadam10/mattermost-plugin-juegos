import React from 'react';

import {TextMove} from '../match/common.jsx';

export default function Pie({match, view, send}) {
    const playing = match.phase === 'escribir' || match.phase === 'votar';
    return (
        <div className='jg-col'>
            {playing && (
                <img
                    className='jg-gif'
                    src={view.gif}
                    alt='GIF de la ronda'
                />
            )}
            {match.phase === 'escribir' && (
                <>
                    <span className='jg-muted'>{`Ronda ${match.round} de ${view.rounds}. Escribe el mejor pie de foto. Llevamos ${view.count}.`}</span>
                    <TextMove
                        initial={view.mine}
                        placeholder='Pie de foto'
                        onSend={(text) => send('move', {text})}
                    />
                </>
            )}
            {match.phase === 'votar' && (
                <>
                    <span className='jg-muted'>{'Vota tu favorito. Son anónimos hasta el final.'}</span>
                    {view.options.map((o, i) => (
                        <button
                            key={i}
                            className={`jg-option ${view.vote === i ? 'jg-chosen' : ''}`}
                            disabled={o.mine}
                            onClick={() => send('move', {option: i})}
                        >
                            {o.mine ? `${o.text} (el tuyo)` : o.text}
                        </button>
                    ))}
                </>
            )}
            {view.reveals.map((r, i) => (
                <details
                    key={i}
                    open={!playing}
                >
                    <summary>{`Ronda ${i + 1}`}</summary>
                    <img
                        className='jg-gif'
                        src={r.gif}
                        alt={`GIF de la ronda ${i + 1}`}
                    />
                    {r.captions.length < 2 && <div className='jg-muted'>{'No hubo pies suficientes para votar.'}</div>}
                    {r.captions.map((c) => (
                        <div key={c.user}>{`${c.text} — @${c.user} · ${c.votes} votos`}</div>
                    ))}
                </details>
            ))}
        </div>
    );
}
