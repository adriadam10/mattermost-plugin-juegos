// Styles follow Mattermost's theme variables, so light and dark themes just work.
export const STYLES = `
.jg-card {
    border: 1px solid rgba(var(--center-channel-color-rgb), 0.16);
    border-radius: 8px;
    padding: 12px 14px;
    margin: 6px 0;
    max-width: 640px;
    background: var(--center-channel-bg);
    color: var(--center-channel-color);
}
.jg-head { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.jg-title { font-weight: 600; font-size: 15px; }
.jg-tag {
    font-size: 11px; padding: 1px 8px; border-radius: 10px;
    background: rgba(var(--center-channel-color-rgb), 0.08);
}
.jg-clock { margin-left: auto; font-variant-numeric: tabular-nums; font-weight: 600; }
.jg-clock.jg-urgent { color: var(--error-text); }
.jg-muted { color: rgba(var(--center-channel-color-rgb), 0.64); font-size: 13px; }
.jg-row { display: flex; gap: 8px; flex-wrap: wrap; align-items: center; margin-top: 10px; }
.jg-col { display: flex; flex-direction: column; gap: 8px; margin-top: 10px; }
.jg-btn {
    border: 0; border-radius: 4px; padding: 6px 14px; font-weight: 600; cursor: pointer;
    background: var(--button-bg); color: var(--button-color);
}
.jg-btn:disabled { opacity: 0.5; cursor: default; }
.jg-btn.jg-ghost {
    background: rgba(var(--button-bg-rgb), 0.08); color: var(--button-bg);
}
.jg-btn.jg-danger { background: transparent; color: var(--error-text); }
.jg-input {
    flex: 1; min-width: 180px; padding: 6px 10px; border-radius: 4px;
    border: 1px solid rgba(var(--center-channel-color-rgb), 0.24);
    background: var(--center-channel-bg); color: var(--center-channel-color);
}
.jg-prompt { font-size: 15px; font-weight: 600; margin: 4px 0; }
.jg-option {
    text-align: left; border-radius: 6px; padding: 8px 12px; cursor: pointer;
    border: 1px solid rgba(var(--center-channel-color-rgb), 0.16);
    background: rgba(var(--center-channel-color-rgb), 0.04); color: inherit;
}
.jg-option:hover:not(:disabled) { border-color: var(--button-bg); }
.jg-option.jg-chosen { border-color: var(--button-bg); background: rgba(var(--button-bg-rgb), 0.12); }
.jg-option:disabled { cursor: default; opacity: 0.7; }
.jg-pair { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; }
.jg-players { display: flex; flex-wrap: wrap; gap: 4px; }
.jg-player {
    font-size: 12px; padding: 2px 8px; border-radius: 10px;
    background: rgba(var(--center-channel-color-rgb), 0.08);
}
.jg-table { border-collapse: collapse; margin-top: 8px; font-size: 13px; }
.jg-table td, .jg-table th { padding: 3px 10px 3px 0; text-align: left; }
.jg-gif { max-width: 100%; max-height: 260px; border-radius: 6px; display: block; margin-top: 8px; }
.jg-numbers { display: grid; grid-template-columns: repeat(10, 1fr); gap: 4px; max-width: 420px; }
.jg-numbers .jg-option { text-align: center; padding: 6px 0; }
.jg-chain { display: flex; gap: 8px; overflow-x: auto; padding: 6px 0; }
.jg-step {
    flex: 0 0 auto; width: 200px; border-radius: 6px; padding: 8px;
    border: 1px solid rgba(var(--center-channel-color-rgb), 0.16);
}
.jg-step img { width: 100%; border-radius: 4px; background: #fff; }
.jg-pad { display: flex; flex-direction: column; gap: 6px; }
.jg-pad canvas {
    width: 100%; max-width: 480px; aspect-ratio: 4 / 3; background: #fff; touch-action: none;
    border-radius: 6px; border: 1px solid rgba(var(--center-channel-color-rgb), 0.24); cursor: crosshair;
}
.jg-swatch {
    width: 22px; height: 22px; border-radius: 50%; cursor: pointer; padding: 0;
    border: 2px solid rgba(var(--center-channel-color-rgb), 0.2);
}
.jg-swatch.jg-chosen { border-color: var(--center-channel-color); transform: scale(1.15); }
.jg-overlay {
    position: fixed; inset: 0; z-index: 1000; background: rgba(0, 0, 0, 0.5);
    display: flex; align-items: center; justify-content: center; padding: 16px;
}
.jg-modal {
    background: var(--center-channel-bg); color: var(--center-channel-color);
    border-radius: 10px; padding: 18px 20px; width: min(640px, 100%);
    max-height: calc(100vh - 32px); overflow-y: auto; box-shadow: 0 12px 32px rgba(0, 0, 0, 0.3);
}
.jg-modal h2 { margin: 0 0 12px; font-size: 20px; }
.jg-games { display: grid; grid-template-columns: repeat(auto-fill, minmax(180px, 1fr)); gap: 8px; }
.jg-game { display: flex; flex-direction: column; gap: 4px; }
.jg-game .jg-emoji { font-size: 24px; }
.jg-canvas-wrap { display: flex; justify-content: center; }
.jg-canvas-wrap canvas {
    width: min(480px, 100%); aspect-ratio: 1; image-rendering: pixelated; cursor: crosshair;
    border: 1px solid rgba(var(--center-channel-color-rgb), 0.24);
}
.jg-error { color: var(--error-text); font-size: 13px; }
`;
