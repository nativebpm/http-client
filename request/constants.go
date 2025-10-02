package request

// HTTP header constants used throughout the request package.
const (
	ApplicationJSON = "application/json"
	ContentType     = "Content-Type"
	ContentLength   = "Content-Length"
)

type DataType int

const (
	NoneType DataType = iota
	ParamType
	FileType
	JSONType
	BytesType
	ReaderType
)
