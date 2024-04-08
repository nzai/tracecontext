package tracecontext

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

type contextKey struct{}

// key is used for context value
var key contextKey

// CtxKeyTid is inserted trace id on the beginning of the request
const (
	XCurrTid  = "X-Curr-Tid"
	XPrevTid  = "X-Prev-Tid"
	XEntryTid = "X-Entry-Tid"
)

var (
	ErrInvalidTraceContext = errors.New("invalid trace context")
)

// TraceContext is inserted to the request context, on the beginning of the request
type TraceContext struct {
	PrevTid  string
	CurrTid  string
	EntryTid string
}

func MustGetTraceContext(ctx context.Context) *TraceContext {
	traceContext, ok := GetTraceContext(ctx)
	if !ok {
		panic("trace context should be inserted into the context, find out why you pass a ctx without it")
	}
	return traceContext
}

func GetTraceContext(ctx context.Context) (traceCtx *TraceContext, ok bool) {
	traceCtx, ok = ctx.Value(key).(*TraceContext)
	ok = ok && traceCtx.IsValid()
	return
}

func (b *TraceContext) IsValid() bool {
	// both entry and curr tid must exist
	return b != nil && b.CurrTid != "" && b.EntryTid != ""
}

func (b *TraceContext) SetHeader(h http.Header) {
	if b.PrevTid != "" {
		h.Set(XPrevTid, b.PrevTid)
	}

	if b.EntryTid == "" {
		h.Set(XEntryTid, b.CurrTid)
	} else {
		h.Set(XEntryTid, b.EntryTid)
	}

	h.Set(XCurrTid, b.CurrTid)
}

func (b *TraceContext) EmbedIntoContext(ctx context.Context) (context.Context, error) {
	if !b.IsValid() {
		return nil, ErrInvalidTraceContext
	}
	return context.WithValue(ctx, key, b), nil
}

func ExtractTraceContextFromHeader(h http.Header) (b TraceContext) {
	b.CurrTid = h.Get(XCurrTid)
	b.PrevTid = h.Get(XPrevTid)
	b.EntryTid = h.Get(XEntryTid)
	return
}

func NewTraceContext() TraceContext {
	tid := NewID()
	return TraceContext{
		EntryTid: tid,
		CurrTid:  tid,
	}
}

func ChainTraceContext(prev *TraceContext) (curr TraceContext, err error) {
	if !prev.IsValid() {
		return curr, ErrInvalidTraceContext
	}
	curr.EntryTid = prev.EntryTid
	curr.PrevTid = prev.CurrTid
	curr.CurrTid = NewID()
	return
}

func newUUID() string {
	return uuid.New().String()
}

var (
	NewID = newUUID
)
