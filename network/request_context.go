package network

import (
	"context"
	"reflect"
)

const REQUEST_CONTEXT = "request_context"

type RequestContext struct {
	szLocalAddress      string
	szRemoteAddress     string
	ptrServerConnection *ServerConnection
	szRequestID         string
}

func GetServerConnection(ctx context.Context) *ServerConnection {
	if ptrMsgContext := GetRequestContext(ctx); ptrMsgContext != nil {
		return ptrMsgContext.ptrServerConnection
	}

	return nil
}

func newContext(ptrServerConnection *ServerConnection, szRequestID string) *RequestContext {
	return &RequestContext{ptrServerConnection.LocalAddr(),
		ptrServerConnection.RemoteAddr(), ptrServerConnection, szRequestID}
}

func GetRequestContext(ctx context.Context) *RequestContext {
	if ptrContext := ctx.Value(REQUEST_CONTEXT); ptrContext != nil {
		if reflect.TypeOf(ptrContext).String() == "*network.RequestContext" {
			return ctx.Value(REQUEST_CONTEXT).(*RequestContext)
		}
	}

	return nil
}

func SetRequestID(ctx context.Context, szRequestID string) {
	if ptrMsgContext := GetRequestContext(ctx); ptrMsgContext != nil {
		ptrMsgContext.szRequestID = szRequestID
	}
}

func GetRequestID(ctx context.Context) string  {
	if ptrMsgContext := GetRequestContext(ctx); ptrMsgContext != nil {
		return ptrMsgContext.szRequestID
	}

	return ""
}

func GetLocalAddress(ctx context.Context) string {
	if ptrMsgContext := GetRequestContext(ctx); ptrMsgContext != nil {
		return ptrMsgContext.szLocalAddress
	}

	return ""
}

func GetRemoteAddress(ctx context.Context) string {
	if ptrMsgContext := GetRequestContext(ctx); ptrMsgContext != nil {
		return ptrMsgContext.szRemoteAddress
	}

	return ""
}
