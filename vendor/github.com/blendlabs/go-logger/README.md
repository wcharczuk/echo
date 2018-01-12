go-logger
=========

`go-logger` is not well named. it is an event bus that can write events in a variety of formats.

Requirements
- Enable or disable event types by flag
- Add one or many listeners for events by flag
- Output can be to one or more writers
- Output can be text (i.e. column based) or json.

Design
- All "Triggers" boil down to an event.
- An "event" is composed of
  - Timestamp
  - Flag (or type?)
  - State
- Most basic messages (info, debug, warning) the state is a string message
- Most error events (error, fatal) the state is an error
- Writing to JSON or writing to Text should look the same to the caller

Example Syntax:

log := logger.NewFromEnv().WithOutputFormat(logger.OutputFormatJSON)

log.Infof("%s foo", bar) 
  - creates a new MessageEvent that is written by an implicit listener to the writer

log.Listen(log.Info, "datadog", func(r logger.Writer, e logger.Event) {

})

