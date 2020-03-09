package actor

// Director of Actors ...
type Director struct {
}

// Direct output from source actor to target actor
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
