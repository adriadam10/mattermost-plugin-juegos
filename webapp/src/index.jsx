import React, {useEffect, useState} from 'react';

import manifest from '../../plugin.json';

import {fetchGames} from './api.js';
import {emit, on} from './bus.js';
import CanvasModal from './canvas.jsx';
import MatchPost from './match/post.jsx';
import Menu from './menu.jsx';
import {STYLES} from './styles.js';
import {localCommand} from './time.js';

// The menu and the canvas are opened from several places (slash command,
// channel header); they live in one root component and are opened through the bus.
function Root() {
    const [open, setOpen] = useState(null);
    useEffect(() => on('open', setOpen), []);
    const close = () => setOpen(null);
    if (open?.what === 'menu') {
        return (
            <Menu
                channelId={open.channelId}
                onClose={close}
                onCanvas={() => setOpen({what: 'canvas'})}
            />
        );
    }
    if (open?.what === 'canvas') {
        return <CanvasModal onClose={close}/>;
    }
    return null;
}

class Plugin {
    initialize(registry, store) {
        this.style = document.createElement('style');
        this.style.textContent = STYLES;
        document.head.appendChild(this.style);

        let trigger = 'juegos';
        fetchGames().then((g) => {
            trigger = g.trigger;
        }, (err) => console.warn(err)); // eslint-disable-line no-console

        const currentChannel = () => store.getState().entities.channels.currentChannelId;
        registry.registerRootComponent(Root);
        registry.registerPostTypeComponent('custom_juegos_match', MatchPost);
        registry.registerChannelHeaderButtonAction(
            <span style={{fontSize: 16}}>{'🎲'}</span>,
            () => emit('open', {what: 'menu', channelId: currentChannel()}),
            'Juegos',
            'Minijuegos',
        );
        registry.registerSlashCommandWillBePostedHook((message, args) => {
            const what = localCommand(message, trigger);
            if (!what) {
                return {message, args};
            }
            emit('open', {what, channelId: args.channel_id});
            return {};
        });
        registry.registerWebSocketEventHandler(`custom_${manifest.id}_match`, (msg) => emit(`match:${msg.data.match_id}`));
        registry.registerWebSocketEventHandler(`custom_${manifest.id}_pixel`, (msg) => emit('pixel', msg.data));
        // Events missed while offline are gone: everything on screen reloads.
        registry.registerReconnectHandler(() => emit('reconnect'));
    }

    uninitialize() {
        this.style?.remove();
    }
}

window.registerPlugin(manifest.id, new Plugin());
