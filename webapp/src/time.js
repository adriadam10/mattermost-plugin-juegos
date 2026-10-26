// Remaining time of a phase as "1:05" under an hour and "3 h 20 min" above.
export function remaining(ms) {
    if (ms <= 0) {
        return '0:00';
    }
    const total = Math.ceil(ms / 1000);
    if (total >= 3600) {
        const h = Math.floor(total / 3600);
        const min = Math.floor((total % 3600) / 60);
        return min ? `${h} h ${min} min` : `${h} h`;
    }
    const min = Math.floor(total / 60);
    const s = String(total % 60).padStart(2, '0');
    return `${min}:${s}`;
}

// The slash command the webapp handles itself, or null: "/juegos" opens the
// menu and "/juegos lienzo" the canvas.
export function localCommand(message, trigger) {
    const words = message.trim().toLowerCase().split(/\s+/);
    if (words[0] !== `/${trigger}`) {
        return null;
    }
    if (words.length === 1) {
        return 'menu';
    }
    if (words.length === 2 && words[1] === 'lienzo') {
        return 'canvas';
    }
    return null;
}
