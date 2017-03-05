package kivik

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

type testMinConfig struct {
	log []string
}

var _ driver.Config = &testMinConfig{}

func (c *testMinConfig) GetAll() (map[string]map[string]string, error) {
	c.log = append(c.log, "GetAll()")
	return map[string]map[string]string{
		"fruit": map[string]string{
			"apple": "red",
		},
	}, nil
}

func (c *testMinConfig) Set(secName, key, value string) error {
	c.log = append(c.log, fmt.Sprintf("Set(%s,%s,%s)", secName, key, value))
	return nil
}

func (c *testMinConfig) Delete(secName, key string) error {
	c.log = append(c.log, fmt.Sprintf("Delete(%s,%s)", secName, key))
	return nil
}

func TestMinimalConfiger(t *testing.T) {
	tc := &testMinConfig{}
	c := &Config{Config: tc}
	_, _ = c.GetAll()
	_ = c.Set("foo", "bar", "baz")
	_ = c.Delete("foo", "bar")
	if _, err := c.GetSection("fruit"); err != nil {
		t.Errorf("Failed to get existing section")
	}
	if _, err := c.GetSection("foo"); errors.StatusCode(err) != http.StatusNotFound {
		t.Errorf("Expected NotFound for non-existant section")
	}
	if _, err := c.Get("fruit", "apple"); err != nil {
		t.Errorf("Failed to get existing value")
	}

	// Existing section, non-existing key
	_, err := c.Get("fruit", "orange")
	if err == nil {
		t.Errorf("Expected NotFound for non-existant key")
	}
	expectedErrMsg := "error status 404: config key not found"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedErrMsg, err.Error())
	}

	// Non-existing section
	_, err = c.Get("animals", "duck")
	if err == nil {
		t.Errorf("Expected NotFound for non-existant key")
	}
	expectedErrMsg = "error status 404: section not found"
	if err.Error() != expectedErrMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedErrMsg, err.Error())
	}
	expected := []string{
		"GetAll()",
		"Set(foo,bar,baz)",
		"Delete(foo,bar)",
		"GetAll()",
		"GetAll()",
		"GetAll()",
		"GetAll()",
		"GetAll()",
	}
	if d := diff.TextSlices(expected, tc.log); d != "" {
		t.Errorf("Log differs:\n%s\n", d)
	}
}

type testSecConfig struct {
	testMinConfig
	log []string
}

func (c *testSecConfig) GetSection(secName string) (map[string]string, error) {
	c.log = append(c.log, fmt.Sprintf("GetSection(%s)", secName))
	return nil, nil
}

func TestSecConfiger(t *testing.T) {
	tc := &testSecConfig{}
	c := &Config{Config: tc}
	_, _ = c.GetSection("fruit")
	_, _ = c.Get("fruit", "salad")
	expectedMinLog := []string{}
	expectedSecLog := []string{
		"GetSection(fruit)",
		"GetSection(fruit)",
	}
	if d := diff.TextSlices(expectedMinLog, tc.testMinConfig.log); d != "" {
		t.Errorf("Min Configer log differs:\n%s\n", d)
	}
	if d := diff.TextSlices(expectedSecLog, tc.log); d != "" {
		t.Errorf("Sec Configer log differs:\n%s\n", d)
	}
}

type testFullConfig struct {
	testSecConfig
	log []string
}

func (c *testFullConfig) Get(secName, key string) (string, error) {
	c.log = append(c.log, fmt.Sprintf("Get(%s,%s)", secName, key))
	return "", nil
}

func TestFullConfiger(t *testing.T) {
	tc := &testFullConfig{}
	c := &Config{Config: tc}
	_, _ = c.GetSection("fruit")
	_, _ = c.Get("fruit", "salad")
	expectedMinLog := []string{}
	expectedSecLog := []string{
		"GetSection(fruit)",
	}
	expectedFullLog := []string{
		"Get(fruit,salad)",
	}
	if d := diff.TextSlices(expectedMinLog, tc.testMinConfig.log); d != "" {
		t.Errorf("Min Configer log differs:\n%s\n", d)
	}
	if d := diff.TextSlices(expectedSecLog, tc.testSecConfig.log); d != "" {
		t.Errorf("Sec Configer log differs:\n%s\n", d)
	}
	if d := diff.TextSlices(expectedFullLog, tc.log); d != "" {
		t.Errorf("Full Configer log differs:\n%s\n", d)
	}
}
