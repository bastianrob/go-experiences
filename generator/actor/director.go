package actor

// Director of Actors ...
type Director struct {
}

// Direct inbox of a target actor, as source actor's outbox
func Direct(actors ...*Actor) {
	var source *Actor
	for _, target := range actors {
		if source == nil {
			source = target
			continue
		}

		source.outbox = target.inbox
		source = target
	}
}
