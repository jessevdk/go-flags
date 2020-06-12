package flags

import (
	"os"
	"testing"
	"time"
)

type options struct {
	TestFlag1  string        `default:"default-value-1" description:"should be auto split"`
	TestFlag2  string        `default:"default-value-2" long:"testflag2" description:"long specified"`
	Test_Flag3 time.Duration `default:"5s" description:"split is specified in name"`
	TestFlag4  int           `default:"4" env:"CUSTOM_FLAG" description:"specify ENV but not long"`
}

func assertOption(t *testing.T, p *Parser, expectedLong string, expectedEnv string, expectedValue interface{}) {
	opt := p.FindOptionByLongName(expectedLong)
	if opt == nil {
		t.Fatalf("unable to find expected options: '%s'", expectedLong)
	}
	if opt.EnvDefaultKey != expectedEnv {
		t.Fatalf("option did not have expected env key '%s'; got '%s'; expected '%s'",
			expectedLong, opt.EnvDefaultKey, expectedEnv)
	}
	if opt.Value() != expectedValue {
		t.Fatalf("option did not have expected value '%s'; got '%#+v'; expected '%#+v'",
			expectedLong, opt.Value(), expectedValue)
	}
}

func assertNoOption(t *testing.T, p *Parser, unexpectedLong string) {
	opt := p.FindOptionByLongName(unexpectedLong)
	if opt != nil {
		t.Fatalf("found option that should not exist: '%s'", unexpectedLong)
	}
}

func TestFullAutoGeneration(t *testing.T) {
	var opts options
	p := NewParser(&opts, Auto)
	left, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("parsing unexpectedly failed: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("didn't expect left over arguments, got %d (%+v)", len(left), left)
	}

	assertOption(t, p, "test_flag_1", "TEST_FLAG_1", "default-value-1")
	assertOption(t, p, "testflag2", "TESTFLAG2", "default-value-2")
	assertOption(t, p, "test_flag3", "TEST_FLAG3", 5*time.Second)
	assertOption(t, p, "test_flag_4", "CUSTOM_FLAG", 4)
}

func TestOnlyLongAutoGeneration(t *testing.T) {
	var opts options
	p := NewParser(&opts, Default|AutoGenerateLong)
	left, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("parsing unexpectedly failed: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("didn't expect left over arguments, got %d (%+v)", len(left), left)
	}

	assertOption(t, p, "test_flag_1", "", "default-value-1")
	assertOption(t, p, "testflag2", "", "default-value-2")
	assertOption(t, p, "test_flag3", "", 5*time.Second)
	assertOption(t, p, "test_flag_4", "CUSTOM_FLAG", 4)
}

func TestOnlyEnvAutoGeneration(t *testing.T) {
	var opts options
	p := NewParser(&opts, Default|AutoGenerateEnv)
	left, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("parsing unexpectedly failed: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("didn't expect left over arguments, got %d (%+v)", len(left), left)
	}

	assertNoOption(t, p, "test_flag_1")
	assertOption(t, p, "testflag2", "TESTFLAG2", "default-value-2")
	assertNoOption(t, p, "test_flag3")
	assertNoOption(t, p, "test_flag_4")
}

func TestFullAutoGenerationWithEnv(t *testing.T) {
	var opts options
	p := NewParser(&opts, Auto)

	os.Setenv("TEST_FLAG_1", "override1")
	os.Setenv("TESTFLAG2", "override2")
	os.Setenv("TEST_FLAG3", "10s")
	os.Setenv("CUSTOM_FLAG", "-1")
	left, err := p.ParseArgs([]string{})
	if err != nil {
		t.Fatalf("parsing unexpectedly failed: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("didn't expect left over arguments, got %d (%+v)", len(left), left)
	}

	assertOption(t, p, "test_flag_1", "TEST_FLAG_1", "override1")
	assertOption(t, p, "testflag2", "TESTFLAG2", "override2")
	assertOption(t, p, "test_flag3", "TEST_FLAG3", 10*time.Second)
	assertOption(t, p, "test_flag_4", "CUSTOM_FLAG", -1)
}

func TestFullAutoGenerationWithFlags(t *testing.T) {
	var opts options
	p := NewParser(&opts, Auto)

	left, err := p.ParseArgs([]string{
		"--test_flag_1=override1-1",
		"--testflag2=override2-2",
		"--test_flag3=15s",
		"--test_flag_4=42",
	})
	if err != nil {
		t.Fatalf("parsing unexpectedly failed: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("didn't expect left over arguments, got %d (%+v)", len(left), left)
	}

	assertOption(t, p, "test_flag_1", "TEST_FLAG_1", "override1-1")
	assertOption(t, p, "testflag2", "TESTFLAG2", "override2-2")
	assertOption(t, p, "test_flag3", "TEST_FLAG3", 15*time.Second)
	assertOption(t, p, "test_flag_4", "CUSTOM_FLAG", 42)
}

func TestFullAutoGenerationWithEnvAndFlags(t *testing.T) {
	var opts options
	p := NewParser(&opts, Auto)

	os.Setenv("TEST_FLAG_1", "override1")
	os.Setenv("TESTFLAG2", "override2")
	os.Setenv("TEST_FLAG3", "10s")
	os.Setenv("CUSTOM_FLAG", "-1")
	left, err := p.ParseArgs([]string{
		"--test_flag_1=override1-1",
		"--testflag2=override2-2",
		"--test_flag3=15s",
		"--test_flag_4=42",
	})
	if err != nil {
		t.Fatalf("parsing unexpectedly failed: %v", err)
	}
	if len(left) != 0 {
		t.Fatalf("didn't expect left over arguments, got %d (%+v)", len(left), left)
	}

	assertOption(t, p, "test_flag_1", "TEST_FLAG_1", "override1-1")
	assertOption(t, p, "testflag2", "TESTFLAG2", "override2-2")
	assertOption(t, p, "test_flag3", "TEST_FLAG3", 15*time.Second)
	assertOption(t, p, "test_flag_4", "CUSTOM_FLAG", 42)
}
