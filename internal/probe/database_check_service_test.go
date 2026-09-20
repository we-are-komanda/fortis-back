package probe

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/valyala/fasthttp"
)

func TestDatabaseReadinessDoesNotAffectLiveness(t *testing.T) {
	calls := 0
	check := NewDatabaseCheckService(func(ctx context.Context) error {
		calls++
		if _, ok := ctx.Deadline(); !ok {
			t.Fatal("DB ping must have a deadline")
		}
		return errors.New("synthetic private connection details")
	})
	controller := NewProbeController(*NewCompositeCheckService(), *NewCompositeCheckService(), *NewCompositeCheckService(check))
	var live, ready fasthttp.RequestCtx
	controller.LivenessProbe(&live)
	if live.Response.StatusCode() != 200 || calls != 0 {
		t.Fatal("liveness must not ping DB")
	}
	controller.ReadinessProbe(&ready)
	if ready.Response.StatusCode() != 500 || calls != 1 {
		t.Fatal("failed DB ping must fail readiness")
	}
	if strings.Contains(string(ready.Response.Body()), "synthetic private connection details") {
		t.Fatal("do not expose connection details")
	}
	if len(NewDatabaseCheckService(func(context.Context) error { return nil }).Check()) != 0 {
		t.Fatal("healthy DB must pass")
	}
}
