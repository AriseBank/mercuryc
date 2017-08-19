package apollo

import (
	"fmt"
	"io/ioutil"
	"os"
	"syscall"
	"testing"
)

func assertNoError(t *testing.T, err error, msg string) {
	if err != nil {
		t.Fatalf("Error: %s, action: %s", err, msg)
	}
}

func TestLocalAPOLLOError(t *testing.T) {
	f, err := ioutil.TempFile("", "apollo-test.socket")
	assertNoError(t, err, "ioutil.TempFile to create fake socket file")
	defer os.RemoveAll(f.Name())

	c := &Client{
		Name:   "test",
		Config: DefaultConfig,
		Remote: &RemoteConfig{
			Addr:   fmt.Sprintf("unix:%s", f.Name()),
			Static: true,
			Public: false,
		},
	}
	runTest := func(exp error) {
		apolloErr := GetLocalAPOLLOErr(connectViaUnix(c, c.Remote))
		if apolloErr != exp {
			t.Fatalf("GetLocalAPOLLOErr returned the wrong error, EXPECTED: %s, ACTUAL: %s", exp, apolloErr)
		}
	}

	// The fake socket file should mimic a socket with nobody listening.
	runTest(syscall.ECONNREFUSED)

	// Remove R/W permissions to mimic the user not having apollo group permissions.
	// Skip this test for root, as root ignores permissions and connect will fail
	// with ECONNREFUSED instead of EACCES.
	if os.Geteuid() != 0 {
		assertNoError(t, f.Chmod(0100), "f.Chmod on fake socket file")
		runTest(syscall.EACCES)
	}

	// Remove the fake socket to mimic APOLLO not being installed.
	assertNoError(t, os.RemoveAll(f.Name()), "osRemoveAll on fake socket file")
	runTest(syscall.ENOENT)
}
