package messages

// Our messages.
var (
	AvailableTasks = "✨  Available Tasks"
	ExistingTask   = "⚠️  task \"%s\" is defined multiple times, using last instance"
	MissingTask    = "🛑  task \"%s\" does not exist"
	StartingTask   = "⚡  %sStarting Task%s \"%s\""
	CompletedTask  = "✔️  %sTask \"%s\" Complete in %s%s"
	FailedTask     = "❌  %sTask \"%s\" Failed%s: %s"
	WatchingTask   = "👀  %sWatching%s"
)
