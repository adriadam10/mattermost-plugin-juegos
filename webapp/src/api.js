import manifest from '../../plugin.json';

// Mattermost rejects cookie-authenticated POSTs without its CSRF header.
function csrfToken() {
    return document.cookie.match(/(?:^|;\s*)MMCSRF=([^;]+)/)?.[1] ?? '';
}

export const apiBase = () => `${window.basename || ''}/plugins/${manifest.id}/api/v1`;

// A refused move (409) resolves with {refused}; anything else that fails throws.
async function request(method, path, body) {
    const response = await fetch(`${apiBase()}/${path}`, {
        method,
        credentials: 'include',
        headers: {
            'Content-Type': 'application/json',
            'X-Requested-With': 'XMLHttpRequest',
            'X-CSRF-Token': csrfToken(),
        },
        body: body && JSON.stringify(body),
    });
    if (response.status === 409) {
        return response.json();
    }
    if (!response.ok) {
        throw new Error(`${manifest.id}: ${method} ${path} -> ${response.status}`);
    }
    return response.json();
}

export const fetchGames = () => request('GET', 'games');
export const createMatch = (game, mode, channelId) => request('POST', 'matches', {game, mode, channel_id: channelId});
export const fetchMatch = (id) => request('GET', `matches/${id}`);
export const act = (id, action, move) => request('POST', `matches/${id}/act`, {action, move});
export const fetchCanvas = () => request('GET', 'canvas');
export const paint = (x, y, color) => request('POST', 'canvas', {x, y, color});
export const fetchRanking = (season) => request('GET', `ranking${season ? `?season=${season}` : ''}`);
export const blobUrl = (key) => `${apiBase()}/blobs/${key}`;
