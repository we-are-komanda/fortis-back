package application

import (
	"context"
	"errors"
	budget "github.com/fortis/backend/internal/modules/budget/domain"
	project "github.com/fortis/backend/internal/modules/defense_project/domain"
	"github.com/fortis/backend/internal/modules/report/domain"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

type projectStub struct {
	p     *project.DefenseProject
	err   error
	calls int
}

func (s *projectStub) GetRevision(_ context.Context, _ string, _ string, v *int) (*project.DefenseProject, error) {
	s.calls++
	if v == nil || *v != 1 {
		panic("unpinned report")
	}
	return s.p, s.err
}

type financeStub struct{ calls int }

func (s *financeStub) FrozenCost(context.Context, *project.DefenseProject) (*budget.CostProjection, error) {
	s.calls++
	return nil, budget.ErrUnsupportedCalculationVersion
}

type repositoryStub struct {
	saved  *domain.IssuedReport
	calls  int
	issued int
	err    error
}

func (s *repositoryStub) FindByKey(context.Context, string, domain.ReportKey) (*domain.IssuedReport, error) {
	s.calls++
	if s.saved == nil {
		return nil, domain.ErrReportNotFound
	}
	return s.saved, nil
}
func (s *repositoryStub) FindByID(context.Context, string, string) (*domain.IssuedReport, error) {
	s.calls++
	return s.saved, s.err
}
func (s *repositoryStub) Issue(_ context.Context, _ string, r *domain.IssuedReport) (*domain.IssuedReport, error) {
	s.issued++
	return r, s.err
}
func TestHiddenReportSkipsFinanceAndReopenSkipsProject(t *testing.T) {
	p, err := project.NewDefenseProject("project", "Variant", "enterprise", "Facility", project.NewProtectedObject("base", "Base", project.NewCoordinates(0, 0)), nil, nil, nil, nil, nil, nil, project.DefenseProjectModeView, project.DefenseProjectSourceCustom, nil, time.Now())
	require.NoError(t, err)
	p.SetVersion(1)
	ps, finance, repo := &projectStub{p: p}, &financeStub{}, &repositoryStub{}
	render := func(d *domain.ReportDraft) (string, error) {
		require.Nil(t, d.Cost())
		require.Nil(t, d.Budget())
		return `{"hideCost":true}`, nil
	}
	service := NewReportService(ps, finance, repo, render)
	key, _ := domain.NewReportKey("project", 1, true, "all")
	got, err := service.GetReport(context.Background(), "actor", GetReportQuery{Key: key})
	require.NoError(t, err)
	require.Zero(t, finance.calls)
	require.Equal(t, 1, repo.issued)
	repo.saved = got
	again, err := service.GetReport(context.Background(), "actor", GetReportQuery{Key: key})
	require.NoError(t, err)
	require.Same(t, got, again)
	require.Equal(t, 1, repo.issued)
	calls := ps.calls
	repo.err = errors.New("revoked")
	_, err = service.GetReport(context.Background(), "actor", GetReportQuery{ReportID: got.ID()})
	require.Equal(t, repo.err, err)
	require.Equal(t, calls, ps.calls)
	require.Zero(t, finance.calls)
}
func TestReportAuthorizesBeforeCacheAndFinance(t *testing.T) {
	denial := errors.New("denied")
	ps, finance, repo := &projectStub{err: denial}, &financeStub{}, &repositoryStub{}
	key, _ := domain.NewReportKey("project", 1, false, "all")
	_, err := NewReportService(ps, finance, repo, nil).GetReport(context.Background(), "actor", GetReportQuery{Key: key})
	require.ErrorIs(t, err, denial)
	require.Zero(t, repo.calls)
	require.Zero(t, finance.calls)
}

func TestPricedIssuanceReusesStoredPayloadWithoutCalculator(t *testing.T) {
	key, _ := domain.NewReportKey("project", 1, false, "all")
	saved := domain.NewIssuedReport("issued", key, "enterprise", time.Now(), `{"immutable":true}`)
	ps, finance, repo := &projectStub{}, &financeStub{}, &repositoryStub{saved: saved}
	got, err := NewReportService(ps, finance, repo, nil).GetReport(context.Background(), "actor", GetReportQuery{Key: key})
	require.NoError(t, err)
	require.Same(t, saved, got)
	require.Zero(t, finance.calls)
	require.Zero(t, repo.issued)
}
