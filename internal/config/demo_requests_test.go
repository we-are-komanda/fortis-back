package config

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDemoRequestsConfigurationDisabledByDefaultAndServerEnvironment(t *testing.T) {
	read := func(string) ([]byte, error) { return []byte("work_mode: webapp\n"), nil }
	cfg := NewConfig("test")
	require.NoError(t, cfg.ReadConfig("unused", read))
	require.False(t, cfg.DemoRequests.Enabled)
	t.Setenv("APP_DEMO_ENABLED", "true")
	t.Setenv("APP_DEMO_CONSENT_VERSIONS", "synthetic-v1,synthetic-v2")
	t.Setenv("APP_DEMO_ALLOWED_ORIGINS", "https://workspace.example.test")
	t.Setenv("APP_DEMO_TRUSTED_PROXY_CIDRS", "127.0.0.0/8,::1/128")
	t.Setenv("APP_DEMO_SMTP_ADDRESS", "smtp.example.test:587")
	t.Setenv("APP_DEMO_SMTP_PASSWORD", "synthetic-only")
	require.NoError(t, cfg.ReadConfig("unused", read))
	require.True(t, cfg.DemoRequests.Enabled)
	require.Equal(t, []string{"synthetic-v1", "synthetic-v2"}, cfg.DemoRequests.ConsentVersions)
	require.Equal(t, []string{"127.0.0.0/8", "::1/128"}, cfg.DemoRequests.TrustedProxyCIDRs)
	require.Equal(t, "smtp.example.test:587", cfg.DemoRequests.SMTPAddress)
	require.Equal(t, "synthetic-only", cfg.DemoRequests.SMTPPassword)
}
