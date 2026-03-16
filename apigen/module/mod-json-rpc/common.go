package mod_json_rpc

import "strings"

const (
	transportName = "JSONRPCHandler"

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

func ignoreModelParam(tmpl, pt, pp string) bool {
	if tmpl == modelNameRes {
		switch {
		case pt == "error":
			return true
		default:
		}
	}

	if tmpl == modelNameReq {
		switch {
		case pp == "context" && pt == "Context":
			return true
		default:
		}
	}

	return false
}

func noBodyParam(vals []string) bool {
	for _, val := range vals {
		i := strings.Index(val, ":")
		if i == -1 {
			continue
		}
		switch strings.ToLower(val[0:i]) {
		case "cookie", "header":
			return true
		default:
		}
	}
	return false
}
