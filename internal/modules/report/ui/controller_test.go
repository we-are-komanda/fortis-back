package ui

import (
	"context"
	"github.com/fortis/backend/internal/modules/report/application"
	"github.com/fortis/backend/internal/modules/report/domain"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"testing"
	"time"
)

type reportStub struct{ queries []application.GetReportQuery }

func (s *reportStub) GetReport(_ context.Context, _ string, q application.GetReportQuery) (*domain.IssuedReport, error) {
	s.queries = append(s.queries, q)
	return domain.NewIssuedReport("report", q.Key, "enterprise", time.Now(), `{"stored":true}`), nil
}
func TestReportQueryModes(t *testing.T) {
	id := uuid.NewString()
	valid := "id=" + id + "&projectVersion=1&hideCost=false"
	invalid := []string{"", "id=" + id, "id=" + id + "&projectVersion=1", valid + "&hideCost=true", valid + "&mapView=", valid + "&mapView=other", valid + "&reportId=" + id, "reportId=bad", "reportId=" + id + "&hideCost=false", valid + "&x=1", "id=" + id + "&projectVersion=0&hideCost=true", "id=" + id + "&projectVersion=x&hideCost=true"}
	for _, hidden := range []string{"1", "TRUE", "False", "garbage", ""} {
		invalid = append(invalid, "id="+id+"&projectVersion=1&hideCost="+hidden)
	}
	for _, query := range invalid {
		t.Run(query, func(t *testing.T) {
			s := &reportStub{}
			var ctx fasthttp.RequestCtx
			ctx.Request.SetRequestURI("/api/v1/projects/report?" + query)
			NewReportController(s).Get(&ctx)
			require.Equal(t, 400, ctx.Response.StatusCode())
			require.Empty(t, s.queries)
		})
	}
	for _, query := range []string{valid, valid + "&mapView=saved", "reportId=" + id} {
		s := &reportStub{}
		var ctx fasthttp.RequestCtx
		ctx.Request.SetRequestURI("/api/v1/projects/report?" + query)
		NewReportController(s).Get(&ctx)
		require.Equal(t, 200, ctx.Response.StatusCode())
		require.JSONEq(t, `{"stored":true}`, string(ctx.Response.Body()))
		require.Len(t, s.queries, 1)
	}
}
