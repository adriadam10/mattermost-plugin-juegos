package main

import (
	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/absurdas"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/cuatro"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/flota"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/guinote"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/numero"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/pie"
	"github.com/adriadam10/mattermost-plugin-juegos/server/games/telefono"
)

// games are the games in the menu, in its order. A new game is a package under
// games/ plus a line here; see the README.
var games = []engine.Game{
	absurdas.Game{},
	pie.Game{},
	telefono.Game{},
	numero.Game{},
	cuatro.Game{},
	flota.Game{},
	guinote.Game{},
}

func gameByID(id string) (engine.Game, bool) {
	for _, g := range games {
		if g.Info().ID == id {
			return g, true
		}
	}
	return nil, false
}
