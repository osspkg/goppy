package mod_json_rpc

const (
	tagWebPool = "web-pool"
)

const (
	transportName = "JSONRPC%sHandle"

	modelNameReq = "jsonrpc%sModelRequest"
	modelNameRes = "jsonrpc%sModelResponse"

	modelBaseReq     = "baseRequest"
	modelBulkBaseReq = "bulkRequest"
	modelBaseRes     = "baseResponse"
	modelBulkBaseRes = "bulkResponse"
	modelBaseErr     = "errResponse"
	errInterface     = "TError"
)

const (
	fieldMethod  = "method"
	fieldParams  = "params"
	fieldID      = "id"
	fieldResult  = "result"
	fieldError   = "error"
	fieldMessage = "message"
	fieldCode    = "code"
	fieldCtx     = "ctx"
)

const (
	jsonGenComment = "easyjson:json"
)
