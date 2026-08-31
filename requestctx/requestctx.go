package requestctx

import (
	"context"

	"github.com/google/uuid"
)

type ctxKey struct{}

// Values is the in-process request identity. Context is the source of truth;
// HTTP headers and gRPC metadata are only used at process boundaries.
type Values struct {
	CorrelationID string
	UserID        string
	TenantID      string
	AgencyID      string
	RequestID     string
	TraceID       string
	SpanID        string
}

func WithValues(ctx context.Context, v Values) context.Context {
	return context.WithValue(ctx, ctxKey{}, v)
}

func From(ctx context.Context) Values {
	if ctx == nil {
		return Values{}
	}
	v, _ := ctx.Value(ctxKey{}).(Values)
	return v
}

func SetUserID(ctx context.Context, userID string) context.Context {
	v := From(ctx)
	v.UserID = userID
	return WithValues(ctx, v)
}

func EnsureIDs(v Values) Values {
	if v.CorrelationID == "" {
		v.CorrelationID = uuid.NewString()
	}
	if v.RequestID == "" {
		v.RequestID = uuid.NewString()
	}
	return v
}

func merge(dst, src Values) Values {
	if dst.CorrelationID == "" {
		dst.CorrelationID = src.CorrelationID
	}
	if dst.UserID == "" {
		dst.UserID = src.UserID
	}
	if dst.TenantID == "" {
		dst.TenantID = src.TenantID
	}
	if dst.AgencyID == "" {
		dst.AgencyID = src.AgencyID
	}
	if dst.RequestID == "" {
		dst.RequestID = src.RequestID
	}
	if dst.TraceID == "" {
		dst.TraceID = src.TraceID
	}
	if dst.SpanID == "" {
		dst.SpanID = src.SpanID
	}
	return dst
}
