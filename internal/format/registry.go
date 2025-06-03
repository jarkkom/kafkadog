package format

// Map of registered codecs
var codecRegistry = make(map[string]func() Codec)

// registerCodec registers a codec factory function
func registerCodec(name string, factory func() Codec) {
	codecRegistry[name] = factory
}

// GetAvailableFormats returns a list of all registered codec formats
func GetAvailableFormats() []string {
	formats := make([]string, 0, len(codecRegistry))
	for format := range codecRegistry {
		formats = append(formats, format)
	}
	return formats
}
