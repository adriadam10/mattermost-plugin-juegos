package pie

import "github.com/adriadam10/mattermost-plugin-juegos/server/engine"

type optionView struct {
	Text string `json:"text"`
	Mine bool   `json:"mine"`
}

type captionView struct {
	Text  string `json:"text"`
	User  string `json:"user"`
	Votes int    `json:"votes"`
}

type revealView struct {
	Gif      string        `json:"gif"`
	Captions []captionView `json:"captions"`
}

func (Game) View(m *engine.Match, userID string, env *engine.Env) any {
	s, err := engine.Load[state](m)
	if err != nil {
		return nil
	}
	r := s.Current
	v := map[string]any{"rounds": s.Rounds, "gif": r.Gif, "count": len(r.Captions)}
	switch m.Phase {
	case phaseWrite:
		v["mine"] = r.Captions[userID]
	case phaseVote:
		options := make([]optionView, len(r.Order))
		vote := -1
		for i, id := range r.Order {
			options[i] = optionView{Text: r.Captions[id], Mine: id == userID}
			if voted, ok := r.Votes[userID]; ok && voted == id {
				vote = i
			}
		}
		v["options"], v["vote"] = options, vote
	}
	reveals := make([]revealView, len(s.Done))
	for i, done := range s.Done {
		tally := done.Votes.Tally()
		reveals[i].Gif = done.Gif
		reveals[i].Captions = []captionView{}
		for _, id := range engine.Ranked(withCaptions(tally, done.Captions)) {
			reveals[i].Captions = append(reveals[i].Captions, captionView{Text: done.Captions[id], User: env.Username(id), Votes: tally[id]})
		}
	}
	v["reveals"] = reveals
	return v
}

func (Game) Notice(m *engine.Match, _ *engine.Env) string {
	switch m.Phase {
	case phaseWrite:
		return "🖼️ GIF nuevo: escribid vuestro pie de foto."
	case phaseVote:
		return "🗳️ Ya se puede votar el mejor pie."
	}
	return ""
}
