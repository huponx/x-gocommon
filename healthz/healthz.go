package healthz

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

type Check struct {
	Name string
	Fn   func(ctx context.Context) error
}

type report struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
}

func Live() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, report{Status: "ok"})
	}
}

func Ready(checks ...Check) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()
		results := map[string]string{}
		ok := true
		for _, check := range checks {
			if err := check.Fn(ctx); err != nil {
				ok = false
				results[check.Name] = err.Error()
				continue
			}
			results[check.Name] = "ok"
		}
		status := http.StatusOK
		body := report{Status: "ok", Checks: results}
		if !ok {
			status = http.StatusServiceUnavailable
			body.Status = "unavailable"
		}
		c.JSON(status, body)
	}
}

// GRPC reports ready when the downstream health service is SERVING.
func GRPC(name string, conn grpc.ClientConnInterface, service string) Check {
	client := healthpb.NewHealthClient(conn)
	return Check{
		Name: name,
		Fn: func(ctx context.Context) error {
			resp, err := client.Check(ctx, &healthpb.HealthCheckRequest{Service: service})
			if err != nil {
				return err
			}
			if resp.GetStatus() != healthpb.HealthCheckResponse_SERVING {
				return errNotServing(resp.GetStatus().String())
			}
			return nil
		},
	}
}

type notServingError string

func (e notServingError) Error() string { return string(e) }

func errNotServing(status string) error {
	return notServingError("health status " + status)
}
