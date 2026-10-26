import React, {useState} from 'react';

import {blobUrl} from '../api.js';
import {TextMove} from '../match/common.jsx';

import DrawPad from './drawpad.jsx';

export default function Telefono({match, view, send}) {
    if (match.phase === 'votar' || match.phase === 'fin') {
        return (
            <Chains
                match={match}
                view={view}
                send={send}
            />
        );
    }
    const mine = view.mine;
    const step = `Paso ${match.round} de ${view.steps} · ${view.done}/${view.steps} listos`;
    if (!mine) {
        return <div className='jg-col jg-muted'>{`${step}. No juegas esta partida; al final podrás votar.`}</div>;
    }
    if (mine.done) {
        return <div className='jg-col jg-muted'>{`${step}. ✅ Enviado. Esperando a los demás…`}</div>;
    }
    return (
        <div className='jg-col'>
            <span className='jg-muted'>{step}</span>
            {match.phase === 'escribir' && (
                <WriteFirst
                    suggestion={view.suggestion}
                    send={send}
                />
            )}
            {match.phase === 'dibujar' && (
                <>
                    <Prev prev={view.prev}/>
                    <Draw send={send}/>
                </>
            )}
            {match.phase === 'describir' && (
                <>
                    <Prev prev={view.prev}/>
                    <TextMove
                        placeholder='¿Qué ves en el dibujo?'
                        onSend={(text) => send('move', {text})}
                    />
                </>
            )}
        </div>
    );
}

function WriteFirst({suggestion, send}) {
    const [initial, setInitial] = useState('');
    return (
        <>
            <span>{'Escribe una frase para que la dibuje el siguiente. Cuanto más rara, mejor.'}</span>
            <TextMove
                initial={initial}
                placeholder='Un pulpo dando una charla TED'
                onSend={(text) => send('move', {text})}
            />
            {suggestion && (
                <button
                    className='jg-btn jg-ghost'
                    style={{alignSelf: 'flex-start'}}
                    onClick={() => setInitial(suggestion)}
                >
                    {'💡 Sugiéreme una'}
                </button>
            )}
        </>
    );
}

function Draw({send}) {
    const [busy, setBusy] = useState(false);
    const onSend = async (drawing) => {
        setBusy(true);
        await send('move', {drawing});
        setBusy(false);
    };
    return (
        <DrawPad
            onSend={onSend}
            busy={busy}
        />
    );
}

function Prev({prev}) {
    if (!prev) {
        return <span className='jg-muted'>{'La cadena llegó vacía: inventa tú.'}</span>;
    }
    if (prev.drawing) {
        return (
            <img
                className='jg-gif'
                style={{background: '#fff'}}
                src={blobUrl(prev.drawing)}
                alt='Dibujo que te ha llegado'
            />
        );
    }
    return <div className='jg-prompt'>{`Dibuja: «${prev.text}»`}</div>;
}

function Chains({match, view, send}) {
    const voting = match.phase === 'votar';
    return (
        <div className='jg-col'>
            {voting && <span className='jg-muted'>{`Vota tu momento favorito (una frase o un dibujo que no sea tuyo). Han votado ${view.voted}.`}</span>}
            {view.chains.map((chain, c) => (
                <div key={c}>
                    <strong>{`Cadena ${c + 1}`}</strong>
                    <div className='jg-chain'>
                        {chain.map((st) => (
                            <div
                                key={st.option}
                                className='jg-step'
                            >
                                {st.drawing ? (
                                    <img
                                        src={blobUrl(st.drawing)}
                                        alt={`Dibujo de ${st.user}`}
                                    />
                                ) : <div>{`«${st.text}»`}</div>}
                                <div className='jg-muted'>{`@${st.user}${voting ? '' : ` · ${st.votes} votos`}`}</div>
                                {voting && !st.mine && (
                                    <button
                                        className={`jg-option ${view.vote === st.option ? 'jg-chosen' : ''}`}
                                        onClick={() => send('move', {vote: st.option})}
                                    >
                                        {view.vote === st.option ? '⭐ Tu voto' : 'Votar'}
                                    </button>
                                )}
                            </div>
                        ))}
                    </div>
                </div>
            ))}
        </div>
    );
}
