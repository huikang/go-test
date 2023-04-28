
makeClientWithConfig

makeClient

NewTestServerConfigT

// snapshot/agent_test.go
TestAgent

// agent/testagent.go
NewTestAgent


```go
func TestAgent_Services() {
	a := NewTestAgent(t, "")
	defer a.Shutdown()
      
	testrpc.WaitForTestAgent(t, a.RPC, "dc1")
	srv1 := &structs.NodeService{
}
```

NewTestAgentWithConfigFile

StartTestAgent
