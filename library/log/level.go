package log

type Level int

const (
	_debugLevel Level = iota
	_infoLevel
	_warnLevel
	_errorLevel
	_fatalLevel
	_accessLevel
	_noLevel
)

var levelName = [...]string{
	_debugLevel:  "DEBUG",
	_infoLevel:   "INFO",
	_warnLevel:   "WARN",
	_errorLevel:  "ERROR",
	_fatalLevel:  "FATAL",
	_accessLevel: "ACCESS",
	_noLevel:     "",
}

func (l Level) String() string {
	return levelName[l]
}
