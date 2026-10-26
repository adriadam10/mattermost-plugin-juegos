import React from 'react';

export default function Numero({match, view, send}) {
    const numbers = Array.from({length: view.max}, (_, i) => i + 1);
    return (
        <div className='jg-col'>
            {match.phase === 'elegir' && (
                <>
                    <span className='jg-muted'>
                        {`Ronda ${match.round} de ${view.rounds}. Elige en secreto: gana el número más bajo que nadie más repita. Puedes cambiarlo hasta el final.`}
                    </span>
                    <div className='jg-numbers'>
                        {numbers.map((n) => (
                            <button
                                key={n}
                                className={`jg-option ${view.picked && view.mine === n ? 'jg-chosen' : ''}`}
                                onClick={() => send('move', {number: n})}
                            >
                                {n}
                            </button>
                        ))}
                    </div>
                    <span className='jg-muted'>{`Han elegido ${view.pickers} (hacen falta ${view.min} para que cuente)`}</span>
                </>
            )}
            {view.done.map((r, i) => (
                <Round
                    key={i}
                    index={i}
                    round={r}
                    min={view.min}
                />
            ))}
        </div>
    );
}

function Round({index, round, min}) {
    const entries = Object.entries(round.picks).sort((a, b) => a[1] - b[1]);
    let result = 'Nadie tuvo un número único.';
    if (entries.length < min) {
        result = 'No cuenta: eligió muy poca gente.';
    } else if (round.winner) {
        result = `Gana @${round.winner} con el ${round.number} 🎉`;
    }
    return (
        <div>
            <strong>{`Ronda ${index + 1}: `}</strong>
            {result}
            <div className='jg-muted'>{entries.map(([u, n]) => `${n} @${u}`).join(' · ')}</div>
        </div>
    );
}
