import assert from 'node:assert/strict';
import {test} from 'node:test';

import {localCommand, remaining} from './time.js';

test('remaining formats minutes and hours', () => {
    assert.equal(remaining(65_000), '1:05');
    assert.equal(remaining(500), '0:01');
    assert.equal(remaining(-1), '0:00');
    assert.equal(remaining(3 * 3600_000 + 20 * 60_000), '3 h 20 min');
    assert.equal(remaining(2 * 3600_000), '2 h');
});

test('localCommand only takes the menu and the canvas', () => {
    assert.equal(localCommand('/juegos', 'juegos'), 'menu');
    assert.equal(localCommand('  /JUEGOS lienzo ', 'juegos'), 'canvas');
    assert.equal(localCommand('/juegos ranking', 'juegos'), null);
    assert.equal(localCommand('/otro', 'juegos'), null);
});
