package meowcaller

import "github.com/rs/zerolog"

// Option configures optional, non-behavioral aspects of the call/media types —
// currently the diagnostic logger. The zero configuration logs nothing.
type Option func(*config)

type config struct {
	log zerolog.Logger
}

func resolveConfig(opts []Option) config {
	c := config{log: zerolog.Nop()}
	for _, opt := range opts {
		opt(&c)
	}
	return c
}

// WithLogger sets the zerolog logger for debug/trace diagnostics. The library never
// configures logging itself; without this option the types are silent at zero cost.
// Pass the logger from a context, e.g. WithLogger(*zerolog.Ctx(ctx)).
func WithLogger(l zerolog.Logger) Option {
	return func(c *config) { c.log = l }
}
