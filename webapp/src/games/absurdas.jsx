import React from 'react';

import {TextMove} from '../match/common.jsx';

export default function Absurdas({match, view, send}) {
    return (
        <div className='jg-col'>
            {match.phase === 'responder' && (
                <>
                    <span className='jg-muted'>{`Responde en secreto. Llevamos ${view.answered} de ${view.total} respuestas.`}</span>
                    {view.tasks.length === 0 && <span className='jg-muted'>{'No juegas esta partida: luego podrás votar.'}</span>}
                    {view.tasks.map((t) => (
                        <div key={t.matchup}>
                            <div className='jg-prompt'>{t.prompt}</div>
                            <TextMove
                                initial={t.answer}
                                placeholder='Tu respuesta más absurda'
                                onSend={(text) => send('move', {matchup: t.matchup, text})}
                            />
                        </div>
                    ))}
                </>
            )}
            {match.phase === 'votar' && (
                <>
                    <span className='jg-muted'>{'Vota la respuesta más graciosa de cada pareja. Son anónimas hasta el final.'}</span>
                    {view.pairs.map((p) => (
                        <div key={p.matchup}>
                            <div className='jg-prompt'>{p.prompt}</div>
                            <div className='jg-pair'>
                                {p.answers.map((a, side) => (
                                    <button
                                        key={side}
                                        className={`jg-option ${p.vote === side ? 'jg-chosen' : ''}`}
                                        disabled={p.mine}
                                        onClick={() => send('move', {matchup: p.matchup, side})}
                                    >
                                        {a}
                                    </button>
                                ))}
                            </div>
                            {p.mine && <span className='jg-muted'>{'Esta es tuya: no puedes votarla.'}</span>}
                        </div>
                    ))}
                </>
            )}
            {view.reveals.length > 0 && (
                <div className='jg-col'>
                    {view.reveals.map((r, i) => (
                        <div key={i}>
                            <div className='jg-prompt'>{r.prompt}</div>
                            {r.answers.map((a, side) => (
                                <div key={side}>
                                    {`${a || '(en blanco)'} — @${r.authors[side]} · ${r.votes[side]} votos · ${r.points[side]} pts`}
                                </div>
                            ))}
                        </div>
                    ))}
                </div>
            )}
        </div>
    );
}
