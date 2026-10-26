import React, {useEffect, useRef, useState} from 'react';

import {fetchCanvas, paint} from './api.js';
import {on} from './bus.js';
import {Overlay} from './match/common.jsx';
import {remaining} from './time.js';

// Pixels and owners arrive as base64 bytes and a plain array.
function decode(base64) {
    return Uint8Array.from(atob(base64), (ch) => ch.charCodeAt(0));
}

export default function CanvasModal({onClose}) {
    const ref = useRef(null);
    const [data, setData] = useState(null);
    const [color, setColor] = useState(1);
    const [hover, setHover] = useState('');
    const [error, setError] = useState('');
    const [, tick] = useState(0);

    const load = (c) => setData({...c, pixels: decode(c.pixels), offset: c.now - Date.now()});
    useEffect(() => {
        fetchCanvas().then(load, () => setError('No se ha podido cargar el lienzo'));
        const id = setInterval(() => tick((n) => n + 1), 1000);
        const offPixel = on('pixel', (p) => setData((d) => {
            if (!d || d.season !== p.season) {
                return d;
            }
            const pixels = d.pixels.slice();
            pixels[(p.y * d.size) + p.x] = p.color;
            return {...d, pixels};
        }));
        return () => {
            clearInterval(id);
            offPixel();
        };
    }, []);

    useEffect(() => {
        if (!data || !ref.current) {
            return;
        }
        const ctx = ref.current.getContext('2d');
        for (let i = 0; i < data.pixels.length; i++) {
            ctx.fillStyle = data.palette[data.pixels[i]];
            ctx.fillRect(i % data.size, Math.floor(i / data.size), 1, 1);
        }
    }, [data]);

    if (!data) {
        return <Overlay onClose={onClose}><p className='jg-muted'>{error || 'Cargando lienzo…'}</p></Overlay>;
    }
    const cell = (e) => {
        const r = ref.current.getBoundingClientRect();
        return [Math.floor(((e.clientX - r.left) / r.width) * data.size), Math.floor(((e.clientY - r.top) / r.height) * data.size)];
    };
    const wait = data.next_at - (Date.now() + data.offset);
    const onClick = async (e) => {
        const [x, y] = cell(e);
        setError('');
        try {
            const res = await paint(x, y, color);
            load(res.canvas);
            if (res.refused) {
                setError(res.refused);
            }
        } catch (err) {
            setError('No se ha podido pintar');
        }
    };
    const onMove = (e) => {
        const [x, y] = cell(e);
        const owner = data.owners[(y * data.size) + x];
        setHover(owner ? `(${x}, ${y}) pintado por @${data.users[owner - 1]}` : `(${x}, ${y})`);
    };

    return (
        <Overlay onClose={onClose}>
            <h2>{`👾 Lienzo de la temporada ${data.season}`}</h2>
            <p className='jg-muted'>{`Un píxel cada ${Math.round(data.cooldown / 60000)} min por persona. A final de mes se publica y empieza uno nuevo.`}</p>
            <div className='jg-canvas-wrap'>
                <canvas
                    ref={ref}
                    width={data.size}
                    height={data.size}
                    onClick={wait > 0 ? undefined : onClick}
                    onMouseMove={onMove}
                    onMouseLeave={() => setHover('')}
                />
            </div>
            <div className='jg-row'>
                {data.palette.map((c, i) => (
                    <button
                        key={c}
                        className={`jg-swatch ${i === color ? 'jg-chosen' : ''}`}
                        style={{background: c}}
                        aria-label={`Color ${c}`}
                        onClick={() => setColor(i)}
                    />
                ))}
            </div>
            <p className='jg-muted'>{wait > 0 ? `⏳ Puedes pintar otro píxel en ${remaining(wait)}` : '✅ Puedes pintar: haz clic en el lienzo.'}</p>
            <p className='jg-muted'>{hover || ' '}</p>
            {error && <p className='jg-error'>{error}</p>}
        </Overlay>
    );
}
