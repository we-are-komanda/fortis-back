package main

import (
	"encoding/json"
	"errors"
	projectDomain "github.com/fortis/backend/internal/modules/defense_project/domain"
	"gorm.io/gorm"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func frc07Document(t *testing.T, enterprise string, price any) string {
	t.Helper()
	var value map[string]any
	decoder := json.NewDecoder(strings.NewReader(frc04Document(enterprise)))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&value))
	value["assetLibrary"].([]any)[0].(map[string]any)["unitPriceMinor"] = price
	object := value["placedObjects"].([]any)[0].(map[string]any)
	delete(object, "customPricePerUnitMln")
	object["quantity"] = 1
	return js(value)
}

func TestFRC07PostgresAtomicBudgetGate(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Budget gate", "enterpriseId": a.e, "projectJson": frc07Document(t, a.e, "47500000")}, 200)
	id := created["projectId"].(string)
	budget := frc04Request(t, f, a, "GET", "/projects/budget?projectId="+id+"&projectVersion=1", "", nil, 200)
	require.Equal(t, "unlimited", budget["status"])
	updated := frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 1, "budgetMode": "limited", "budgetLimitMinor": "50000000"}, 200)
	require.EqualValues(t, 2, updated["projectVersion"])
	require.NotNil(t, updated["snapshot"])
	count := func(table string) int64 { var n int64; require.NoError(t, db.Table(table).Count(&n).Error); return n }
	before := f.request(t, a, "GET", "/projects/get?id="+id, nil, 200)
	revisions, events, caches := count("project_revisions"), count("audit_events"), count("project_cost_projections")
	check := frc04Request(t, f, a, "POST", "/projects/budget/check?projectId="+id, "", map[string]any{"version": 2, "projectJson": frc07Document(t, a.e, "50000001")}, 200)
	require.Equal(t, "blocked", check["status"])
	require.Equal(t, "budget_exceeded", check["reasonCode"])
	blocked := frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 2, "projectJson": frc07Document(t, a.e, "50000001")}, 422)
	require.Contains(t, js(blocked), "budget_exceeded")
	require.Equal(t, before, f.request(t, a, "GET", "/projects/get?id="+id, nil, 200))
	require.Equal(t, revisions, count("project_revisions"))
	require.Equal(t, events, count("audit_events"))
	require.Equal(t, caches, count("project_cost_projections"))
	frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 2, "projectJson": frc07Document(t, a.e, "50000000")}, 200)
	current := frc04Request(t, f, a, "GET", "/projects/budget?projectId="+id+"&projectVersion=3", "", nil, 200)
	require.Equal(t, "0", current["remainingMinor"])
	// Explicit lowering may create overage; it still records a full revision.
	lowered := frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 3, "budgetMode": "limited", "budgetLimitMinor": "40000000"}, 200)
	require.Equal(t, "exceeded", lowered["budget"].(map[string]any)["status"])
	check = frc04Request(t, f, a, "POST", "/projects/budget/check?projectId="+id, "", map[string]any{"version": 4, "projectJson": frc07Document(t, a.e, "50000000")}, 200)
	require.Equal(t, "non_increasing_overage", check["reasonCode"])
	frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 4, "projectJson": frc07Document(t, a.e, "50000001")}, 422)
	frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 4, "projectJson": frc07Document(t, a.e, "45000000")}, 200)
	historical := frc04Request(t, f, a, "GET", "/projects/budget?projectId="+id+"&projectVersion=2", "", nil, 200)
	require.Equal(t, "50000000", historical["config"].(map[string]any)["budgetLimitMinor"])
}

func TestFRC07PostgresConcurrentBudgetGate(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Concurrent budget", "enterpriseId": a.e, "projectJson": frc07Document(t, a.e, "47500000")}, 200)
	id := created["projectId"].(string)
	frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 1, "budgetMode": "limited", "budgetLimitMinor": "50000000"}, 200)
	service := frc04ProjectService(db)
	start := make(chan struct{})
	failures := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			version := 2
			_, err := service.UpdateProject(t.Context(), a.u, id, "", nil, frc07Document(t, a.e, "49500000"), &version)
			failures <- err
		}()
	}
	close(start)
	wg.Wait()
	close(failures)
	success, conflict := 0, 0
	for err := range failures {
		if err == nil {
			success++
		} else if errors.Is(err, projectDomain.ErrVersionConflict) {
			conflict++
		} else {
			t.Fatalf("unexpected writer error: %v", err)
		}
	}
	require.Equal(t, 1, success)
	require.Equal(t, 1, conflict)
	frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 3, "projectJson": frc07Document(t, a.e, "51500000")}, 422)
}

func TestFRC07PostgresUnknownZeroAndValidation(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a, b := f.people[0], f.people[1]
	created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Unknown cost", "enterpriseId": a.e, "projectJson": frc07Document(t, a.e, nil)}, 200)
	id := created["projectId"].(string)
	projection := frc04Request(t, f, a, "GET", "/projects/budget?projectId="+id, "", nil, 200)
	require.Equal(t, "indeterminate", projection["status"])
	require.Nil(t, projection["spentMinor"])
	unknown := frc07Document(t, a.e, nil)
	check := frc04Request(t, f, a, "POST", "/projects/budget/check?projectId="+id, "", map[string]any{"version": 1, "projectJson": unknown}, 200)
	require.Equal(t, "allowed", check["status"])
	require.Equal(t, "unlimited", check["reasonCode"])
	require.Nil(t, check["projectedSpentMinor"])
	updated := frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 1, "budgetMode": "limited", "budgetLimitMinor": "0"}, 200)
	require.Equal(t, "indeterminate", updated["budget"].(map[string]any)["status"])
	check = frc04Request(t, f, a, "POST", "/projects/budget/check?projectId="+id, "", map[string]any{"version": 2, "projectJson": unknown}, 200)
	require.Equal(t, "price_unknown", check["reasonCode"])
	require.Nil(t, check["projectedSpentMinor"])
	rejected := frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 2, "projectJson": unknown}, 422)
	details := rejected["error"].(map[string]any)["details"].(map[string]any)
	require.Equal(t, "placedObjects", details["field"])
	require.NotEmpty(t, details["objectIds"])
	require.Nil(t, details["currentSpentMinor"])
	// A known over-limit candidate cannot prove non-increase from an unknown baseline.
	rejected = frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 2, "projectJson": frc07Document(t, a.e, "1")}, 422)
	require.Equal(t, "budget_exceeded", rejected["error"].(map[string]any)["code"])
	frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 2, "projectJson": frc07Document(t, a.e, "0")}, 200)
	zero := frc04Request(t, f, a, "GET", "/projects/budget?projectId="+id, "", nil, 200)
	require.Equal(t, "within_limit", zero["status"])
	require.Equal(t, "0", zero["spentMinor"])
	for _, limit := range []any{nil, 1, "", "-1", "1.5", "NaN", "Infinity", "1e2"} {
		frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 3, "budgetMode": "limited", "budgetLimitMinor": limit}, 400)
	}
	frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 3, "budgetMode": "limited"}, 400)
	frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 0, "budgetMode": "unlimited", "budgetLimitMinor": nil}, 400)
	frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 2, "budgetMode": "unlimited", "budgetLimitMinor": nil}, 409)
	frc04Request(t, f, a, "POST", "/projects/budget/check?projectId="+id, "", map[string]any{"version": 2, "projectJson": unknown}, 409)
	frc04Request(t, f, a, "POST", "/projects/budget/check?projectId="+id, "", map[string]any{"version": 3, "projectJson": frc07Document(t, b.e, "0")}, 400)
	for _, route := range []string{"/projects/budget?projectId=" + id + "&id=" + b.p, "/projects/budget?projectId=" + id + "&projectVersion=0"} {
		frc04Request(t, f, a, "GET", route, "", nil, 400)
	}
	frc04Request(t, f, a, "GET", "/projects/budget?projectId="+id+"&projectVersion=999", "", nil, 404)
	for _, request := range []struct {
		method, path string
		body         any
	}{
		{"GET", "/projects/budget?projectId=" + id + "&projectVersion=2", nil},
		{"PUT", "/projects/budget?projectId=" + id, map[string]any{"version": 2, "budgetMode": "limited", "budgetLimitMinor": "1"}},
		{"POST", "/projects/budget/check?projectId=" + id, map[string]any{"version": 2, "projectJson": unknown}},
	} {
		want := 200
		if request.method != "GET" {
			want = 409
		}
		frc04Request(t, f, b, request.method, request.path, "", request.body, want)
	}
	frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 3, "budgetMode": "unlimited", "budgetLimitMinor": nil}, 200)
	frc04Request(t, f, a, "PUT", "/projects/update?id="+id, "", map[string]any{"version": 4, "projectJson": unknown}, 200)
}

func TestFRC07PostgresCloneAuthorityAndReplay(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a, b := f.people[0], f.people[1]
	source := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Source", "enterpriseId": a.e, "projectJson": frc07Document(t, a.e, "10")}, 200)
	id := source["projectId"].(string)
	frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 1, "budgetMode": "limited", "budgetLimitMinor": "10"}, 200)
	var forged map[string]any
	require.NoError(t, json.Unmarshal([]byte(frc07Document(t, a.e, "10")), &forged))
	forged["budget"] = map[string]any{"budgetMode": "limited", "budgetLimitMinor": "0"}
	forged["budgetMode"], forged["budgetLimitMinor"] = "limited", "0"
	payload := map[string]any{"name": "Clone", "enterpriseId": a.e, "projectJson": js(forged), "budgetSource": map[string]any{"projectId": id, "projectVersion": 2}}
	cloned := frc04Request(t, f, a, "POST", "/projects", "clone-budget", payload, 200)
	require.NotEqual(t, id, cloned["projectId"])
	require.Equal(t, "10", cloned["snapshot"].(map[string]any)["budget"].(map[string]any)["budgetLimitMinor"])
	require.Equal(t, cloned, frc04Request(t, f, a, "POST", "/projects", "clone-budget", payload, 200))
	// Source later changes, but a pinned v2 clone still inherits ten kopecks.
	frc04Request(t, f, a, "PUT", "/projects/budget?projectId="+id, "", map[string]any{"version": 2, "budgetMode": "limited", "budgetLimitMinor": "5"}, 200)
	pinned := frc04Request(t, f, a, "POST", "/projects", "clone-pinned", payload, 200)
	require.Equal(t, "10", pinned["snapshot"].(map[string]any)["budget"].(map[string]any)["budgetLimitMinor"])
	payload["budgetSource"] = map[string]any{"projectId": id, "projectVersion": 3}
	frc04Request(t, f, a, "POST", "/projects", "clone-budget", payload, 409)
	// A new clone has no existing-overage baseline; rejected reservation rolls back.
	frc04Request(t, f, a, "POST", "/projects", "rejected-clone", payload, 422)
	var count int64
	require.NoError(t, db.Table("project_idempotency").Where("idempotency_key = ?", "rejected-clone").Count(&count).Error)
	require.Zero(t, count)
	payload["projectJson"] = frc07Document(t, a.e, "5")
	repaired := frc04Request(t, f, a, "POST", "/projects", "rejected-clone", payload, 200)
	require.Equal(t, "5", repaired["snapshot"].(map[string]any)["budget"].(map[string]any)["budgetLimitMinor"])
	// New/import ignore every embedded budget and import also ignores top-level source.
	delete(payload, "budgetSource")
	payload["projectJson"] = js(forged)
	fresh := frc04Request(t, f, a, "POST", "/projects", "", payload, 200)
	require.Equal(t, "unlimited", fresh["snapshot"].(map[string]any)["budget"].(map[string]any)["budgetMode"])
	imported := frc04Request(t, f, a, "POST", "/projects/import", "", map[string]any{"projectJson": js(forged), "budgetSource": map[string]any{"projectId": b.p, "projectVersion": 1}}, 200)
	require.Equal(t, "unlimited", imported["snapshot"].(map[string]any)["budget"].(map[string]any)["budgetMode"])
	payload["budgetSource"] = map[string]any{"projectId": b.p, "projectVersion": 1}
	frc04Request(t, f, a, "POST", "/projects", "", payload, 200)
	payload["budgetSource"] = map[string]any{"projectId": id, "projectVersion": 999}
	frc04Request(t, f, a, "POST", "/projects", "", payload, 404)
	payload["budgetSource"] = map[string]any{"projectId": id, "projectVersion": 0}
	frc04Request(t, f, a, "POST", "/projects", "", payload, 400)
}

func TestFRC07PostgresRollbackAndBudgetRace(t *testing.T) {
	db := frc04Database(t)
	f := newPostgresFixture(t, db)
	a := f.people[0]
	created := frc04Request(t, f, a, "POST", "/projects", "", map[string]any{"name": "Atomic changes", "enterpriseId": a.e, "projectJson": frc07Document(t, a.e, "100")}, 200)
	id := created["projectId"].(string)
	count := func(table string) int64 { var n int64; require.NoError(t, db.Table(table).Count(&n).Error); return n }
	readBudget := func() string {
		var row struct{ ConfigData string }
		require.NoError(t, db.Table("budget_configs").Where("project_id = ?", id).Take(&row).Error)
		return row.ConfigData
	}
	for _, table := range []string{"budget_configs", "project_revisions", "audit_events"} {
		t.Run("rollback_"+table, func(t *testing.T) {
			before := f.request(t, a, "GET", "/projects/get?id="+id, nil, 200)
			beforeBudget := readBudget()
			revisions, events := count("project_revisions"), count("audit_events")
			hook := "frc07_failure_" + table
			require.NoError(t, db.Callback().Create().Before("gorm:create").Register(hook, func(tx *gorm.DB) {
				if tx.Statement.Table == table {
					tx.AddError(errors.New("synthetic transactional failure"))
				}
			}))
			_, err := frc04ProjectService(db).UpdateBudget(t.Context(), a.u, id, 1, `{"budgetMode":"limited","budgetLimitMinor":"50"}`)
			require.ErrorContains(t, err, "synthetic transactional failure")
			require.NoError(t, db.Callback().Create().Remove(hook))
			require.Equal(t, before, f.request(t, a, "GET", "/projects/get?id="+id, nil, 200))
			require.Equal(t, beforeBudget, readBudget())
			require.Equal(t, revisions, count("project_revisions"))
			require.Equal(t, events, count("audit_events"))
		})
	}
	// Budget and composition updates contend for the SAME version/row lock.
	service := frc04ProjectService(db)
	start := make(chan struct{})
	results := make(chan error, 2)
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		<-start
		_, err := service.UpdateBudget(t.Context(), a.u, id, 1, `{"budgetMode":"limited","budgetLimitMinor":"50"}`)
		results <- err
	}()
	go func() {
		defer wg.Done()
		<-start
		v := 1
		_, err := service.UpdateProject(t.Context(), a.u, id, "", nil, frc07Document(t, a.e, "150"), &v)
		results <- err
	}()
	close(start)
	wg.Wait()
	close(results)
	success, conflict := 0, 0
	for err := range results {
		if err == nil {
			success++
		} else if errors.Is(err, projectDomain.ErrVersionConflict) {
			conflict++
		} else {
			t.Fatalf("unexpected race error: %v", err)
		}
	}
	require.Equal(t, 1, success)
	require.Equal(t, 1, conflict)
	current := frc04Request(t, f, a, "GET", "/projects/get?id="+id, "", nil, 200)
	require.EqualValues(t, 2, current["version"])
	var revisions int64
	require.NoError(t, db.Table("project_revisions").Where("project_id = ?", id).Count(&revisions).Error)
	require.EqualValues(t, 2, revisions)
	// The response snapshot and separately read budget always agree after either winner.
	projection := frc04Request(t, f, a, "GET", "/projects/budget?projectId="+id, "", nil, 200)
	require.Equal(t, current["snapshot"].(map[string]any)["budget"], projection["config"])
}
