package signals

// The dispatcher stops on the first non-nil error.
//
// Slots may eventually return io.EOF when they are closed. For now, `eof` is kept as nil so
// the dispatcher doesn't need a special-case check (e.g., errors.Is(err, io.EOF)).
//
// This is a marker for a later modification.
var eof error = nil
