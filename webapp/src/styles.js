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
.jg-turn { font-weight: 600; color: var(--online-indicator); }
.jg-c4 { display: grid; gap: 4px; max-width: 360px; padding: 6px; border-radius: 8px; background: #1f4fbf; }
.jg-c4-cell {
    aspect-ratio: 1; border: 0; border-radius: 50%; padding: 0; font-size: 26px; line-height: 1;
    background: var(--center-channel-bg); cursor: pointer;
}
.jg-c4-cell:disabled { cursor: default; }
.jg-c4-line { box-shadow: 0 0 0 3px #fff, 0 0 8px 3px #ffec27; }
.jg-c4-last { box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.8); }
.jg-boards { display: flex; gap: 16px; flex-wrap: wrap; }
.jg-grid { display: grid; gap: 1px; width: min(280px, 100%); margin-top: 4px; }
.jg-sea {
    aspect-ratio: 1; border: 0; padding: 0; font-size: 12px; line-height: 1; cursor: pointer;
    background: rgba(41, 173, 255, 0.35); color: var(--center-channel-color);
}
.jg-sea:disabled { cursor: default; }
.jg-sea:hover:not(:disabled) { background: rgba(41, 173, 255, 0.7); }
.jg-sea.jg-water { background: rgba(41, 173, 255, 0.15); }
.jg-sea.jg-ship { background: rgba(var(--center-channel-color-rgb), 0.55); }
.jg-sea.jg-last { outline: 2px solid var(--button-bg); outline-offset: -2px; }
.jg-mesa {
    position: relative; display: grid; grid-template-columns: 1fr 2fr 1fr; grid-template-rows: auto auto auto;
    gap: 6px; padding: 10px; border-radius: 10px; background: #1d6b3a; color: #fff;
}
.jg-mesa .jg-muted { color: rgba(255, 255, 255, 0.75); }
.jg-seat { display: flex; flex-direction: column; align-items: center; font-size: 13px; padding: 4px; border-radius: 6px; }
.jg-seat.jg-on-turn { background: rgba(255, 255, 255, 0.18); }
.jg-seat-top { grid-column: 2; grid-row: 1; }
.jg-seat-left { grid-column: 1; grid-row: 2; justify-content: center; }
.jg-seat-right { grid-column: 3; grid-row: 2; justify-content: center; }
.jg-seat-bottom { grid-column: 2; grid-row: 3; }
.jg-center {
    grid-column: 2; grid-row: 2; display: grid; grid-template-columns: repeat(3, 1fr); grid-template-rows: repeat(3, auto);
    justify-items: center; align-items: center; min-height: 150px;
}
.jg-slot-2 { grid-column: 2; grid-row: 1; }
.jg-slot-3 { grid-column: 1; grid-row: 2; }
.jg-slot-1 { grid-column: 3; grid-row: 2; }
.jg-slot-0 { grid-column: 2; grid-row: 3; }
.jg-muestra { grid-column: 3; grid-row: 3; display: flex; flex-direction: column; align-items: center; gap: 2px; font-size: 12px; }
.jg-muestra .jg-naipe { transform: rotate(90deg); margin: 8px 0; }
.jg-naipe { width: 46px; border-radius: 4px; box-shadow: 0 1px 3px rgba(0, 0, 0, 0.4); display: block; }
.jg-naipe-s { width: 32px; }
.jg-dim { opacity: 0.4; }
.jg-hand { display: flex; gap: 6px; flex-wrap: wrap; justify-content: center; }
.jg-hand .jg-naipe { width: 64px; }
.jg-naipe-btn { border: 0; padding: 0; background: none; cursor: pointer; transition: transform 0.1s; }
.jg-naipe-btn:hover { transform: translateY(-6px); }
.jg-doom { width: min(960px, 100%); }
.jg-doom iframe { width: 100%; aspect-ratio: 4 / 3; border: 0; border-radius: 6px; background: #000; }
`;
