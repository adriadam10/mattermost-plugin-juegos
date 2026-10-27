// Fleet placement rules, as on the server (server/games/flota/fleet.go): ships
// inside the board, not touching each other, not even at a corner.

export function shipAt(ships, r, c) {
    return ships.some((s) => s.some(([sr, sc]) => sr === r && sc === c));
}

export function validPlacement(ships, size) {
    const taken = new Map();
    for (let i = 0; i < ships.length; i++) {
        for (const [r, c] of ships[i]) {
            if (r < 0 || r >= size || c < 0 || c >= size) {
                return false;
            }
            for (let dr = -1; dr <= 1; dr++) {
                for (let dc = -1; dc <= 1; dc++) {
                    const owner = taken.get(`${r + dr},${c + dc}`);
                    if (owner !== undefined && owner !== i) {
                        return false;
                    }
                }
            }
            taken.set(`${r},${c}`, i);
        }
    }
    return true;
}

export function randomFleet(fleet, size, random = Math.random) {
    for (;;) {
        const ships = [];
        for (const n of fleet) {
            for (let tries = 0; tries < 200; tries++) {
                const across = random() < 0.5;
                const r = Math.floor(random() * size);
                const c = Math.floor(random() * size);
                const ship = Array.from({length: n}, (_, i) => (across ? [r, c + i] : [r + i, c]));
                if (validPlacement([...ships, ship], size)) {
                    ships.push(ship);
                    break;
                }
            }
        }
        if (ships.length === fleet.length) {
            return ships;
        }
    }
}
