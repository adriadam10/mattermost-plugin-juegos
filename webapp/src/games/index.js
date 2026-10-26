import Absurdas from './absurdas.jsx';
import Numero from './numero.jsx';
import Pie from './pie.jsx';
import Telefono from './telefono.jsx';

// Board of each game by id, as in server/registry.go. A component gets
// {match, view, send}: view is what the game's View returns on the server, and
// send(action, move) posts a move and resolves to false if it was refused.
export const GAMES = {
    absurdas: Absurdas,
    numero: Numero,
    pie: Pie,
    telefono: Telefono,
};
