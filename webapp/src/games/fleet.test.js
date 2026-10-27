import assert from 'node:assert/strict';
import {test} from 'node:test';

import {randomFleet, validPlacement} from './fleet.js';

const FLEET = [4, 3, 3, 2, 2, 2, 1, 1, 1, 1];

test('random fleets follow the rules', () => {
    for (let i = 0; i < 50; i++) {
        const ships = randomFleet(FLEET, 10);
        assert.equal(ships.length, FLEET.length);
        assert.ok(validPlacement(ships, 10));
    }
});

test('ships touching at a corner or outside are refused', () => {
    assert.equal(validPlacement([[[0, 0], [0, 1]], [[1, 2]]], 10), false);
    assert.equal(validPlacement([[[0, 9], [0, 10]]], 10), false);
    assert.equal(validPlacement([[[0, 0], [0, 1]], [[2, 2]]], 10), true);
});
