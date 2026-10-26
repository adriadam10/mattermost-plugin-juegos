import React, {useEffect, useRef, useState} from 'react';

const WIDTH = 480;
const HEIGHT = 360;
const COLORS = ['#000000', '#ffffff', '#ff004d', '#ffa300', '#ffec27', '#00e436', '#29adff', '#7e2553', '#ab5236', '#83769c'];
const SIZES = [3, 8, 20];

// DrawPad is a small paint program: colours, three brush sizes, undo and clear.
// It hands back a PNG data URL.
export default function DrawPad({onSend, busy}) {
    const canvas = useRef(null);
    const history = useRef([]);
    const last = useRef(null);
    const [color, setColor] = useState(COLORS[0]);
    const [size, setSize] = useState(SIZES[1]);
    const [dirty, setDirty] = useState(false);

    useEffect(() => clear(), []);

    const ctx = () => canvas.current.getContext('2d');
    function clear() {
        const c = ctx();
        c.fillStyle = '#ffffff';
        c.fillRect(0, 0, WIDTH, HEIGHT);
        history.current = [];
        setDirty(false);
    }
    const point = (e) => {
        const r = canvas.current.getBoundingClientRect();
        return [(e.clientX - r.left) * (WIDTH / r.width), (e.clientY - r.top) * (HEIGHT / r.height)];
    };
    const line = (from, to) => {
        const c = ctx();
        c.strokeStyle = color;
        c.lineWidth = size;
        c.lineCap = 'round';
        c.lineJoin = 'round';
        c.beginPath();
        c.moveTo(...from);
        c.lineTo(...to);
        c.stroke();
    };
    const down = (e) => {
        canvas.current.setPointerCapture(e.pointerId);
        history.current.push(ctx().getImageData(0, 0, WIDTH, HEIGHT));
        last.current = point(e);
        line(last.current, last.current);
        setDirty(true);
    };
    const move = (e) => {
        if (!last.current) {
            return;
        }
        const p = point(e);
        line(last.current, p);
        last.current = p;
    };
    const up = () => {
        last.current = null;
    };
    const undo = () => {
        const prev = history.current.pop();
        if (prev) {
            ctx().putImageData(prev, 0, 0);
        }
        setDirty(history.current.length > 0);
    };

    return (
        <div className='jg-pad'>
            <canvas
                ref={canvas}
                width={WIDTH}
                height={HEIGHT}
                onPointerDown={down}
                onPointerMove={move}
                onPointerUp={up}
                onPointerCancel={up}
            />
            <div className='jg-row'>
                {COLORS.map((c) => (
                    <button
                        key={c}
                        className={`jg-swatch ${c === color ? 'jg-chosen' : ''}`}
                        style={{background: c}}
                        aria-label={`Color ${c}`}
                        onClick={() => setColor(c)}
                    />
                ))}
            </div>
            <div className='jg-row'>
                {SIZES.map((s) => (
                    <button
                        key={s}
                        className={`jg-option ${s === size ? 'jg-chosen' : ''}`}
                        onClick={() => setSize(s)}
                    >
                        <span style={{display: 'inline-block', width: s, height: s, borderRadius: '50%', background: 'currentColor', verticalAlign: 'middle'}}/>
                    </button>
                ))}
                <button
                    className='jg-btn jg-ghost'
                    onClick={undo}
                    disabled={!dirty}
                >
                    {'↶ Deshacer'}
                </button>
                <button
                    className='jg-btn jg-ghost'
                    onClick={clear}
                    disabled={!dirty}
                >
                    {'Borrar'}
                </button>
                <button
                    className='jg-btn'
                    disabled={!dirty || busy}
                    onClick={() => onSend(canvas.current.toDataURL('image/png'))}
                >
                    {'Enviar dibujo'}
                </button>
            </div>
        </div>
    );
}
