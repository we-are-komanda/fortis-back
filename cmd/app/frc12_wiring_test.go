package main

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestFRC12PublicRouteDisabledAndNoRegistry(t *testing.T) {
	f := newAccessFixture(t)
	body := f.request(t, person{}, "POST", "/demo-requests", map[string]any{}, 503)
	require.Contains(t, body, "unavailable")
	f.request(t, person{}, "GET", "/demo-requests/registry", nil, 401)
	body = f.request(t, f.people[0], "GET", "/demo-requests/registry", nil, 404)
	require.False(t, strings.Contains(body, "email"))
}
