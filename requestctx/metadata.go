package requestctx

import (
	"context"

	"google.golang.org/grpc/metadata"
)

const (
	MDCorrelationID = "correlation-id"
	MDUserID        = "user-id"
	MDTenantID      = "tenant-id"
	MDAgencyID      = "agency-id"
	MDRequestID     = "request-id"
	MDTraceID       = "trace-id"
	MDSpanID        = "span-id"
)

func FromMD(md metadata.MD) Values {
	if md == nil {
		return Values{}
	}
	first := func(key string) string {
		vals := md.Get(key)
		if len(vals) == 0 {
			return ""
		}
		return vals[0]
	}
	return Values{
		CorrelationID: first(MDCorrelationID),
		UserID:        first(MDUserID),
		TenantID:      first(MDTenantID),
		AgencyID:      first(MDAgencyID),
		RequestID:     first(MDRequestID),
		TraceID:       first(MDTraceID),
		SpanID:        first(MDSpanID),
	}
}

func HydrateIncoming(ctx context.Context) context.Context {
	md, _ := metadata.FromIncomingContext(ctx)
	v := merge(From(ctx), FromMD(md))
	return WithValues(ctx, EnsureIDs(v))
}

// AppendOutgoing copies Values onto outgoing metadata without overwriting keys
// the caller already set.
func AppendOutgoing(ctx context.Context) context.Context {
	v := From(ctx)
	md, ok := metadata.FromOutgoingContext(ctx)
	if ok {
		md = md.Copy()
	} else {
		md = metadata.MD{}
	}
	setIfAbsent := func(key, val string) {
		if val == "" || len(md.Get(key)) > 0 {
			return
		}
		md.Set(key, val)
	}
	setIfAbsent(MDCorrelationID, v.CorrelationID)
	setIfAbsent(MDUserID, v.UserID)
	setIfAbsent(MDTenantID, v.TenantID)
	setIfAbsent(MDAgencyID, v.AgencyID)
	setIfAbsent(MDRequestID, v.RequestID)
	setIfAbsent(MDTraceID, v.TraceID)
	setIfAbsent(MDSpanID, v.SpanID)
	return metadata.NewOutgoingContext(ctx, md)
}
