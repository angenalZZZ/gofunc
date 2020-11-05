package log

// CronLogger used by Cron.
type CronLogger struct {
	Log *Logger
}

// Info logs routine messages about cron's operation.
func (c *CronLogger) Info(msg string, keysAndValues ...interface{}) {
	c.Log.Info().Msgf(msg, keysAndValues...)
}

// Error logs an error condition.
func (c *CronLogger) Error(err error, msg string, keysAndValues ...interface{}) {
	c.Log.Error().Err(err).Msgf(msg, keysAndValues)
}
