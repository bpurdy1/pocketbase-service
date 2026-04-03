package logging

import "io"

func WithLevel(level string) Option {
	return func(c *option) {
		c.LogLevel = level
	}
}

func WithConsole() Option {
	return func(c *option) {
		c.ConsoleWriter = true
	}
}

func WithShortCaller() Option {
	return func(c *option) {
		c.CallerMarshalFunc = ShortCallerMarshalFunc
	}
}

func WithFileBaseCaller() Option {
	return func(c *option) {
		c.CallerMarshalFunc = FileBaseCallerMarshalFunc
	}
}

func WithWriter(w io.Writer) Option {
	return func(c *option) {
		c.Writer = w
	}
}
