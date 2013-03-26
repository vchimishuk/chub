package config

import (
	"testing"
)

const testConfigFilename = "test.conf"

func assertString(t *testing.T, key, expected, actual string) {
	if expected != actual {
		t.Fatalf("%s key %s value expected but %s received",
			key, expected, actual)
	}
}

func TestPackage(t *testing.T) {
	conf, err := parse(testConfigFilename)
	if err != nil {
		t.Fatalf("Failed to parse file. %s", err)
	}

	assertString(t, outputName, "alsa", conf.OutputName())
	assertString(t, vfsRoot, getHomeDir(), conf.VfsRoot())
}
