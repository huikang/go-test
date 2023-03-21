package fromunittest

import (
	"testing"

	"github.com/hashicorp/consul/sdk/testutil"
	"github.com/hashicorp/consul/sdk/testutil/retry"
	"github.com/stretchr/testify/require"
)

func TestAPI_HealthChecks(t *testing.T) {
	t.Parallel()
	nodename := "node123"
	c, s := makeClientWithConfig(t, nil, func(conf *testutil.TestServerConfig) {
		conf.NodeName = nodename
	})
	defer s.Stop()

	agent := c.Agent()
	health := c.Health()

	// Make a service with a check
	reg := &AgentServiceRegistration{
		Name: "foo",
		Tags: []string{"bar"},
		Check: &AgentServiceCheck{
			TTL: "15s",
		},
	}
	if err := agent.ServiceRegister(reg); err != nil {
		t.Fatalf("err: %v", err)
	}

	retry.Run(t, func(r *retry.R) {
		checks := HealthChecks{
			&HealthCheck{
				Node:        nodename,
				CheckID:     "service:foo",
				Name:        "Service 'foo' check",
				Status:      "critical",
				ServiceID:   "foo",
				ServiceName: "foo",
				ServiceTags: []string{"bar"},
				Type:        "ttl",
				Partition:   defaultPartition,
				Namespace:   defaultNamespace,
			},
		}

		out, meta, err := health.Checks("foo", nil)
		if err != nil {
			r.Fatal(err)
		}
		if meta.LastIndex == 0 {
			r.Fatalf("bad: %v", meta)
		}
		checks[0].CreateIndex = out[0].CreateIndex
		checks[0].ModifyIndex = out[0].ModifyIndex
		require.Equal(r, checks, out)
	})
}

func TestAPI_HealthChecks_NodeMetaFilter(t *testing.T) {
	t.Parallel()
	meta := map[string]string{"somekey": "somevalue"}
	c, s := makeClientWithConfig(t, nil, func(conf *testutil.TestServerConfig) {
		conf.NodeMeta = meta
	})
	defer s.Stop()

	agent := c.Agent()
	health := c.Health()

	s.WaitForSerfCheck(t)

	// Make a service with a check
	reg := &AgentServiceRegistration{
		Name: "foo",
		Check: &AgentServiceCheck{
			TTL: "15s",
		},
	}
	if err := agent.ServiceRegister(reg); err != nil {
		t.Fatalf("err: %v", err)
	}

	retry.Run(t, func(r *retry.R) {
		checks, meta, err := health.Checks("foo", &QueryOptions{NodeMeta: meta})
		if err != nil {
			r.Fatal(err)
		}
		if meta.LastIndex == 0 {
			r.Fatalf("bad: %v", meta)
		}
		if len(checks) != 1 {
			r.Fatalf("expected 1 check, got %d", len(checks))
		}
		if checks[0].Type != "ttl" {
			r.Fatalf("expected type ttl, got %s", checks[0].Type)
		}
	})
}
