package fromunittest

func TestAgent_Services(t *testing.T) {
	if testing.Short() {
		t.Skip("too slow for testing.Short")
	}

	t.Parallel()
	a := NewTestAgent(t, "")
	defer a.Shutdown()

	testrpc.WaitForTestAgent(t, a.RPC, "dc1")
	srv1 := &structs.NodeService{
		ID:      "mysql",
		Service: "mysql",
		Tags:    []string{"primary"},
		Meta: map[string]string{
			"foo": "bar",
		},
		Port: 5000,
	}
	require.NoError(t, a.State.AddServiceWithChecks(srv1, nil, "", false))

	req, _ := http.NewRequest("GET", "/v1/agent/services", nil)
	resp := httptest.NewRecorder()
	a.srv.h.ServeHTTP(resp, req)

	require.Equal(t, http.StatusOK, resp.Code)
	decoder := json.NewDecoder(resp.Body)
	var val map[string]*api.AgentService
	err := decoder.Decode(&val)
	require.NoError(t, err)
	assert.Lenf(t, val, 1, "bad services: %v", val)
	assert.Equal(t, 5000, val["mysql"].Port)
	assert.Equal(t, srv1.Meta, val["mysql"].Meta)
}
