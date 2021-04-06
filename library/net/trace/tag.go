package trace

const (
	// The software package, framework, library, or module that generated the associated Span.
	// eg. `grpc`, `django`, `JDBI`.
	// type string
	TagComponent = "component"

	// Database instance name.
	// eg. In java, if the jdbc.url=`jdbc:mysql://127.0.0.1:3306/customers`, the instance name is `customers`.
	// type string
	TagDBInstance = "db.instance"

	// A database statement for the given database type.
	// eg. for db.type="sql", "select * from user_table"; fro db.type="redis", "set key 'value'".
	// type string
	TagDbStatement = "db.statement"

	// type bool
	TagError = "error"

	// HTTP method of the request for the associated Span. eg, "GET", "POST"
	// type string
	TagHTTPMethod = "http.method"

	// HTTP response status code for the associated Span. eg, 200, 503, 504
	// type int
	TagHTTPStatusCode = "http.status_code"

	// URL of the request being handled in this segment of the trace, in standard URI format.
	// eg. "https://domain.net/path/to?resource=here"
	// type string
	TagHTTPURL = "http.url"

	// Remote "address", suitable for use in a networking client library.
	// This may be a "ip:port", a bare "hostname", a FQDN, or even a JDBC substring like "mysql://prod-db:3306"
	// type string
	TagPeerAddress = "peer.address"

	// Remote service name (for some unspecified definition of "service").
	// Eg, "elasticsearch", "a_custom_microservice", "memcache"
	// type string
	TagPeerService = "peer.service"

	// Either "client" or "server" for the appropriate roles in an RPC, and
	// "producer" or "consumer" for the appropriate roles in a messaging scenario.
	// type string
	TagSpanKind = "span.kind"
)

const (
	LogErrorKind   = "error.kind"
	LogErrorObject = "error.object"
	LogEvent       = "event"
	LogMessage     = "message"
	LogStack       = "stack"
)

type Tag struct {
	Key   string
	Value interface{}
}

func TagString(key, val string) Tag {
	return Tag{Key: key, Value: val}
}

func TagInt64(key string, val int64) Tag {
	return Tag{Key: key, Value: val}
}

func TagInt(key string, val int) Tag {
	return Tag{Key: key, Value: val}
}

func TagBool(key string, val bool) Tag {
	return Tag{Key: key, Value: val}
}

func TagFloat64(key string, val float64) Tag {
	return Tag{Key: key, Value: val}
}

func TagFloat32(key string, val float64) Tag {
	return Tag{Key: key, Value: val}
}

type LogField struct {
	Key   string
	Value string
}

func Log(key string, val string) LogField {
	return LogField{Key: key, Value: val}
}
