package client

// NewMeta creates a new meta object.
func NewMeta() *Meta {
	return &Meta{
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
}

// Meta is optional extra metadata for a message.
type Meta struct {
	Labels      map[string]string
	Annotations map[string]string
}

// WithLabel adds a label to the metadata.
// Labels are values that are used for filtering messages
// and have some strict requirements around the format and length
// of both the keys and the values.
func (m *Meta) WithLabel(key, value string) *Meta {
	if m.Labels == nil {
		m.Labels = map[string]string{}
	}
	m.Labels[key] = value
	return m
}

// WithAnnotation adds an annotation to the metadata.
// They are not used for filtering and have looser requirements
// around the keys and the values.
func (m *Meta) WithAnnotation(key, value string) *Meta {
	if m.Annotations == nil {
		m.Annotations = map[string]string{}
	}
	m.Annotations[key] = value
	return m
}
