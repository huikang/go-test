package testdata

func pred() bool {
	return true
}

// func pp(x int) int {
// 	if x > 2 && pred() {
// 		return 5
// 	}

// 	var b = pred()
// 	if b {
// 		return 6
// 	}

// 	pred()
// 	pp2()
// 	return 0
// }

func pp0() {
	y = y + 1
}

func pp2() {
	y := 5
	if x > 2 && pred() {
		x = x + 1
		return 5
	}
	pred()
}

func pp3() {
	pred()
}

// func makeClientWithConfig(
// 	t *testing.T,
// 	cb1 configCallback,
// 	cb2 testutil.ServerConfigCallback) (*Client, *testutil.TestServer) {
// 	return
// }

func a() {}
