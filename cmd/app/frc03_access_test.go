package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFRC03ConfigurationNameHTTP(t *testing.T) {
	f := newAccessFixture(t)
	actor := f.people[0]
	before := len(f.projects.projects)
	for _, name := range []string{"", " \t ", strings.Repeat("я", 121), strings.Repeat("😀", 121)} {
		f.request(t, actor, "POST", "/projects", map[string]any{"name": name, "enterpriseId": actor.e, "projectJson": frc04Document(actor.e)}, 400)
		f.request(t, actor, "PUT", "/projects/update?id="+actor.p, map[string]any{"name": name, "version": 1}, 400)
		f.request(t, actor, "POST", "/projects/import", map[string]any{"projectJson": strings.Replace(frc04Document(actor.e), `"projectName":`, `"name":`+js(name)+`,"projectName":`, 1)}, 400)
	}
	require.Len(t, f.projects.projects, before)
	require.Equal(t, 1, f.projects.projects[actor.p].Version())
}
