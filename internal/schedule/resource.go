package schedule

// Resource represents a named resource with a capacity for scheduling.
type Resource struct {
	Name     string
	Capacity int
	events   []Event
}

// NewResource creates a resource with the given capacity.
func NewResource(name string, capacity int) *Resource {
	return &Resource{Name: name, Capacity: capacity}
}

// Schedule attempts to schedule an event on this resource. Returns false if
// the event would exceed capacity.
func (r *Resource) Schedule(e Event) bool {
	// Count how many events overlap at the peak of the new event.
	overlap := 0
	for _, existing := range r.events {
		if existing.ToInterval().Overlaps(e.ToInterval()) {
			overlap++
		}
	}
	if overlap >= r.Capacity {
		return false
	}
	r.events = append(r.events, e)
	return true
}

// Events returns all scheduled events.
func (r *Resource) Events() []Event {
	out := make([]Event, len(r.events))
	copy(out, r.events)
	return out
}

// Load returns the number of scheduled events.
func (r *Resource) Load() int {
	return len(r.events)
}

// PeakLoad returns the maximum number of concurrent events.
func (r *Resource) PeakLoad() int {
	return MaxOverlap(r.events)
}

// Available reports whether an event can be scheduled without exceeding capacity.
func (r *Resource) Available(e Event) bool {
	overlap := 0
	for _, existing := range r.events {
		if existing.ToInterval().Overlaps(e.ToInterval()) {
			overlap++
		}
	}
	return overlap < r.Capacity
}

// Clear removes all scheduled events.
func (r *Resource) Clear() {
	r.events = nil
}
