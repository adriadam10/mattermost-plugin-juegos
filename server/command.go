package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/plugin"

	"github.com/adriadam10/mattermost-plugin-juegos/server/engine"
)

func command(trigger string) *model.Command {
	auto := model.NewAutocompleteData(trigger, "", "Abre el menú de minijuegos")
	for _, g := range games {
		info := g.Info()
		sub := model.NewAutocompleteData(info.ID, "[directo|async]", info.Name+": "+info.Description)
		sub.AddStaticListArgument("Modo", false, []model.AutocompleteListItem{
			{Item: string(engine.Live), HelpText: "Fases de un minuto"},
			{Item: string(engine.Async), HelpText: "Fases de horas, entre tarea y tarea"},
		})
		auto.AddCommand(sub)
	}
	auto.AddCommand(model.NewAutocompleteData("lienzo", "", "Abre el lienzo de la temporada"))
	auto.AddCommand(model.NewAutocompleteData("ranking", "[AAAA-MM]", "Ranking de la temporada"))
	auto.AddCommand(model.NewAutocompleteData("frase", "<tipo> <texto> | quitar <tipo> <texto>", "Añade una frase o un tema de GIF a los juegos"))
	auto.AddCommand(model.NewAutocompleteData("ayuda", "", "Cómo se usa"))
	return &model.Command{
		Trigger:          trigger,
		DisplayName:      "Juegos",
		Description:      "Minijuegos para el equipo",
		AutoComplete:     true,
		AutoCompleteDesc: "Abre el menú de minijuegos",
		AutocompleteData: auto,
	}
}

func ephemeral(text string) *model.CommandResponse {
	return &model.CommandResponse{ResponseType: model.CommandResponseTypeEphemeral, Text: text}
}

func (p *Plugin) ExecuteCommand(_ *plugin.Context, args *model.CommandArgs) (*model.CommandResponse, *model.AppError) {
	trigger := p.settings.Load().trigger
	fields := strings.Fields(args.Command)
	sub := ""
	if len(fields) > 1 {
		sub = strings.ToLower(fields[1])
	}
	switch sub {
	// The webapp opens the menu and the canvas itself; this answers the mobile apps.
	case "", "ayuda", "lienzo":
		return ephemeral(help(trigger)), nil
	case "ranking":
		season := p.season(time.Now())
		if len(fields) > 2 {
			season = fields[2]
		}
		text, err := p.rankingText(season, 20)
		if err != nil {
			return nil, model.NewAppError("ExecuteCommand", "juegos.ranking", nil, err.Error(), 500)
		}
		return ephemeral(text), nil
	case "frase":
		return ephemeral(p.fraseCommand(fields[2:], args.UserId, trigger)), nil
	}
	if _, ok := gameByID(sub); !ok {
		return ephemeral(fmt.Sprintf("No conozco «%s».\n\n%s", sub, help(trigger))), nil
	}
	mode := engine.Live
	if len(fields) > 2 {
		mode = engine.Mode(strings.ToLower(fields[2]))
	}
	if _, err := p.createMatch(args.UserId, args.ChannelId, sub, mode); err != nil {
		if engine.IsRefusal(err) {
			return ephemeral(err.Error()), nil
		}
		p.API.LogError("Failed to create a match", "game", sub, "err", err.Error())
		return ephemeral("No se ha podido crear la partida: " + err.Error()), nil
	}
	return &model.CommandResponse{}, nil
}

func (p *Plugin) fraseCommand(args []string, userID, trigger string) string {
	remove := len(args) > 0 && args[0] == "quitar"
	if remove {
		args = args[1:]
	}
	if len(args) < 2 {
		return fraseHelp(trigger)
	}
	kind, text := strings.ToLower(args[0]), strings.Join(args[1:], " ")
	var err error
	if remove {
		err = p.removeContent(kind, text, userID)
	} else {
		err = p.addContent(kind, text, userID)
	}
	if err != nil {
		return "❌ " + err.Error()
	}
	if remove {
		return "🗑️ Quitada."
	}
	return "✅ Añadida. Saldrá en las próximas partidas."
}

func fraseHelp(trigger string) string {
	kinds := make([]string, 0, len(contentKinds))
	for k := range contentKinds {
		kinds = append(kinds, k)
	}
	sort.Strings(kinds)
	var b strings.Builder
	fmt.Fprintf(&b, "Uso: `/%s frase <tipo> <texto>`. Tipos:\n", trigger)
	for _, k := range kinds {
		fmt.Fprintf(&b, "- `%s`: %s\n", k, contentKinds[k])
	}
	fmt.Fprintf(&b, "\nEjemplo: `/%s frase absurdas Lo que de verdad significa «está casi»: ___`", trigger)
	return b.String()
}

func help(trigger string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "🎲 **Minijuegos.** En el navegador o el escritorio, `/%s` abre el menú. También:\n", trigger)
	for _, g := range games {
		info := g.Info()
		fmt.Fprintf(&b, "- `/%s %s [directo|async]`: %s %s\n", trigger, info.ID, info.Emoji, info.Name)
	}
	fmt.Fprintf(&b, "- `/%s lienzo`: el lienzo de píxeles de la temporada\n", trigger)
	fmt.Fprintf(&b, "- `/%s ranking [AAAA-MM]`: ranking de la temporada\n", trigger)
	fmt.Fprintf(&b, "- `/%s frase`: añade frases a los juegos\n", trigger)
	b.WriteString("\nLas partidas se juegan en el navegador o el escritorio; en el móvil solo se ven los resultados.")
	return b.String()
}
