package requestctx

import "net/http"

const (
	HTTPHeaderCorrelationID = "X-Correlation-ID"
	HTTPHeaderRequestID     = "X-Request-ID"
	HTTPHeaderTenantID      = "X-Tenant-ID"
	HTTPHeaderAgencyID      = "X-Agency-ID"
	HTTPHeaderTraceID       = "X-Trace-ID"
	HTTPHeaderSpanID        = "X-Span-ID"
)

// FromHTTP reads inbound identity headers. UserID is intentionally omitted:
// resolve it from auth and call SetUserID.
func FromHTTP(r *http.Request) Values {
	if r == nil {
		return EnsureIDs(Values{})
	}
	return EnsureIDs(Values{
		CorrelationID: r.Header.Get(HTTPHeaderCorrelationID),
		TenantID:      r.Header.Get(HTTPHeaderTenantID),
		AgencyID:      r.Header.Get(HTTPHeaderAgencyID),
		RequestID:     r.Header.Get(HTTPHeaderRequestID),
		TraceID:       r.Header.Get(HTTPHeaderTraceID),
		SpanID:        r.Header.Get(HTTPHeaderSpanID),
	})
}

func ToHTTPHeaders(h http.Header, v Values) {
	if h == nil {
		return
	}
	set := func(key, val string) {
		if val != "" {
			h.Set(key, val)
		}
	}
	set(HTTPHeaderCorrelationID, v.CorrelationID)
	set(HTTPHeaderRequestID, v.RequestID)
	set(HTTPHeaderTenantID, v.TenantID)
	set(HTTPHeaderAgencyID, v.AgencyID)
	set(HTTPHeaderTraceID, v.TraceID)
	set(HTTPHeaderSpanID, v.SpanID)
}
