package apperror

import (
	"net/http"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestFromGRPCStatus(t *testing.T) {
	err := status.Error(codes.Unavailable, "downstream down")
	got := From(err)
	if got.Code != codes.Unavailable {
		t.Fatalf("code: %v", got.Code)
	}
	if HTTPStatus(err) != http.StatusBadGateway {
		t.Fatalf("http: %d", HTTPStatus(err))
	}
}

func TestFromAppError(t *testing.T) {
	err := NotFound("user missing")
	if HTTPStatus(err) != http.StatusNotFound {
		t.Fatalf("http: %d", HTTPStatus(err))
	}
	st := status.Convert(err)
	if st.Code() != codes.NotFound {
		t.Fatalf("grpc: %v", st.Code())
	}
}
