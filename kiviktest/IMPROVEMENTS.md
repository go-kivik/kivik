# kiviktest: Gradual Improvement Roadmap

Constraint: minimum Go version is 1.20 (`go.mod`).

---

## Batch 4: Add `c.DB()` helper (boilerplate reduction)

Add helper methods to `kt.Context`:

```go
func (c *Context) DB(t *testing.T, client *kivik.Client, dbname string) *kivik.DB
func (c *Context) AdminDB(t *testing.T, dbname string) *kivik.DB
```

Then update call sites across `client/` and `db/` test files.

---

## Batch 5: SuiteConfig redesign

`SuiteConfig map[string]any` uses stringly-typed keys with hierarchical
dot-separated lookup. Research into prior art suggests alternatives:

- **Go CDK drivertest**: typed `Harness` interface per driver, no config map
- **K8s Gateway API conformance**: feature flags as typed constants with
  `SupportedFeatures` sets
- **SQLAlchemy dialect compliance**: `Requirements` class with typed
  skip-predicates per capability
- **W3C WPT / Level ecosystem**: metadata files declaring expected
  pass/fail per test per backend

Possible directions:
1. Typed capabilities struct replacing `map[string]any`
2. Harness interface where drivers declare supported operations
3. Hybrid: keep config map for error expectations, add typed fields for
   feature flags and skip conditions

This is orthogonal to the T-extraction (Batch 3, now complete) and can
be planned separately.
