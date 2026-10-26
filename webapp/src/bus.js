// A tiny event bus between the websocket handlers and the components on screen.
const listeners = new Map();

export function on(topic, fn) {
    if (!listeners.has(topic)) {
        listeners.set(topic, new Set());
    }
    listeners.get(topic).add(fn);
    return () => listeners.get(topic)?.delete(fn);
}

export function emit(topic, payload) {
    for (const fn of listeners.get(topic) ?? []) {
        fn(payload);
    }
}
