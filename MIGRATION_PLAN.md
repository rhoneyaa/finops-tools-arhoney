# Executive Summary Migration Plan

## Overview
Migrate the executive summary report functionality from HCMFinOpsResources-v2 (Python) to finops-tools (Go), preserving business logic while adapting to Go idioms and the current repository structure.

## Source Code Analysis

### Legacy Structure (Python)
**Location:** `/Users/aaronrhoney/Desktop/HCMFinOpsResources-v2/finops-core`

```
finops-core/
├── src/finops_core/
│   ├── reports/
│   │   ├── executive_summary.py      # Core analytics functions
│   │   └── monthly_savings.py
│   ├── transports/
│   │   └── executive_summary.py      # I/O validation & data loading
│   ├── analytics/
│   │   ├── exec_summary_math.py
│   │   └── helpers/
│   │       ├── aggregation.py
│   │       ├── results.py
│   │       └── windows.py
│   ├── catalog/
│   │   └── reader.py
│   ├── queries/
│   │   └── builder.py
│   └── schema/
```

### Key Python Modules

#### 1. `reports/executive_summary.py` (PRIMARY)
**Core analytics functions:**
- `month_label(date) -> str` - Format date as YYYY-MM
- `month_name(date) -> str` - Format date as "Mon YYYY"
- `months_in_window(start, end) -> list[date]` - Generate month range
- `compute_window(n_months, today) -> (start, end)` - Calculate report window
- `enrich_with_mapping(cost_df, mapping_df) -> DataFrame` - Join cost data with account metadata
- `category_monthly(cost_df, months) -> DataFrame` - Aggregate by category/month
- `top_growing_accounts(cost_df, months, top_n, top_svcs) -> list[dict]` - MoM growth analysis
- `statistical_anomalies(cost_df, months, z_threshold) -> list[dict]` - Z-score anomaly detection
- `compute_payer_kpis(...) -> dict` - Compute all KPIs for a payer

**Data Dependencies:**
- Pandas DataFrames with columns: `month`, `account_id`, `cost`, `service`, `payer`
- Account mapping with: `account_id_clean`, `Refined Category`, `Sub Type`, `Owner Team`, `Account Name`
- HCP accounts set
- Cluster counts dict
- Environment counts dict
- Savings Plan data (coverage, utilization)

#### 2. `transports/executive_summary.py`
**I/O and validation:**
- `load_cost_transport(path) -> DataFrame` - Load and validate cost CSV/JSON
- `load_sp_transport(path) -> dict` - Load savings plan data
- `load_cluster_counts(csv_path) -> dict` - Load cluster count timeseries
- `load_env_cluster_counts(csv_path) -> dict` - Load live environment counts
- Validation functions for JSON/CSV transports

#### 3. `analytics/exec_summary_math.py`
Additional analytics helpers (needs investigation)

## Target Structure (Go)

### Current finops-tools Layout
```
finops-tools/
├── core/
│   ├── account/          # Account management
│   ├── cost/             # Cost retrieval (AWS CE)
│   └── report/           # Report data assembly
│       ├── costs.go
│       └── progress.go
├── cli/
│   ├── cmd/              # Cobra commands
│   ├── internal/
│   │   ├── output/       # Human-readable formatting
│   │   ├── report/       # HTML templates & charts
│   │   └── configstore/  # YAML config
```

### Proposed Migration Structure

```
core/report/
├── execsummary/
│   ├── analytics.go      # Core analytics (month_label, compute_window, etc.)
│   ├── enrichment.go     # enrich_with_mapping, category_monthly
│   ├── anomalies.go      # top_growing_accounts, statistical_anomalies
│   ├── kpis.go           # compute_payer_kpis
│   ├── types.go          # Data structures (CostRecord, AccountMapping, KPIs)
│   └── loader.go         # Data loading (replaces transports)
```

```
cli/internal/
├── cmd/
│   └── report_generate.go     # Extended for exec-summary template
├── report/
│   └── templates/
│       └── executive-summary.html.tmpl  # Jinja2 template
```

## Data Structure Mapping

### Python → Go

#### Cost Data
```python
# Python (Pandas DataFrame)
cost_df.columns = ['month', 'account_id', 'cost', 'service', 'payer']
```

```go
// Go
type CostRecord struct {
    Month     string  `json:"month"`      // YYYY-MM
    AccountID string  `json:"account_id"`
    Cost      float64 `json:"cost"`
    Service   string  `json:"service,omitempty"`
    Payer     string  `json:"payer,omitempty"`
}

type CostData struct {
    Records []CostRecord
    // Index maps for efficient lookups
    ByMonth     map[string][]CostRecord
    ByAccount   map[string][]CostRecord
    ByMonthAcct map[string]map[string]float64
}
```

#### Account Mapping
```python
# Python
mapping_df.columns = ['account_id_clean', 'Refined Category', 'Sub Type', 'Owner Team', 'Account Name']
```

```go
// Go
type AccountMapping struct {
    AccountID        string `json:"account_id"`
    RefinedCategory  string `json:"refined_category"`
    SubType          string `json:"sub_type"`
    OwnerTeam        string `json:"owner_team"`
    AccountName      string `json:"account_name"`
}
```

#### KPIs
```python
# Python dict
{
    "total_last": float,
    "total_prev": float,
    "mom_pct": float,
    ...
}
```

```go
// Go
type PayerKPIs struct {
    TotalLast        float64              `json:"total_last"`
    TotalPrev        float64              `json:"total_prev"`
    MoMPct           float64              `json:"mom_pct"`
    HCPCost          float64              `json:"hcp_cost"`
    HCPUnit          *float64             `json:"hcp_unit,omitempty"`
    ClusterCount     int                  `json:"cluster_count"`
    ClusterDelta     *int                 `json:"cluster_delta,omitempty"`
    AnomalyCount     int                  `json:"anomaly_count"`
    Anomalies        []AnomalyRecord      `json:"anomalies"`
    SPCoverage       []SavingsPlanMetric  `json:"sp_coverage"`
    SPUtilization    []SavingsPlanMetric  `json:"sp_utilization"`
}
```

## Migration Strategy

### Phase 1: Core Analytics (Week 1)
**Files:** `core/report/execsummary/analytics.go`, `types.go`

1. ✅ Port date/time utilities:
   - `MonthLabel(time.Time) string`
   - `MonthName(time.Time) string`
   - `MonthsInWindow(start, end time.Time) []time.Time`
   - `ComputeWindow(nMonths int, today time.Time) (start, end time.Time)`

2. ✅ Define core types:
   - `CostRecord`, `CostData`
   - `AccountMapping`
   - `AnomalyRecord`, `GrowingAccountRecord`

### Phase 2: Data Loading & Validation (Week 1-2)
**Files:** `core/report/execsummary/loader.go`

1. Implement CSV/JSON loaders:
   - `LoadCostTransport(path string) (*CostData, error)`
   - `LoadSavingsPlansTransport(path string) (*SPData, error)`
   - `LoadClusterCounts(path string) (map[string]int, error)`
   - `LoadEnvClusterCounts(path string) (map[string]interface{}, error)`

2. Add validation:
   - Non-empty file checks
   - Required columns validation
   - Data integrity checks

### Phase 3: Enrichment & Aggregation (Week 2)
**Files:** `core/report/execsummary/enrichment.go`

1. Port enrichment logic:
   - `EnrichWithMapping(costData *CostData, mappings []AccountMapping) error`
   - `CategoryMonthly(costData *CostData, months []time.Time) []CategoryMonthRecord`

2. Implement aggregation helpers:
   - Month-based filtering
   - Account grouping
   - Service grouping

### Phase 4: Anomaly Detection (Week 2-3)
**Files:** `core/report/execsummary/anomalies.go`

1. Port growth analysis:
   - `TopGrowingAccounts(costData *CostData, months []time.Time, topN, topSvcs int) []GrowingAccountRecord`
   - Month-over-month delta calculation
   - Top services per account

2. Port statistical anomalies:
   - `StatisticalAnomalies(costData *CostData, months []time.Time, zThreshold float64) []AnomalyRecord`
   - Z-score calculation
   - Historical mean/stddev

### Phase 5: KPI Computation (Week 3)
**Files:** `core/report/execsummary/kpis.go`

1. Implement payer KPI aggregation:
   - `ComputePayerKPIs(enriched *CostData, payerLabel string, ...) (*PayerKPIs, error)`
   - Total cost calculations
   - HCP cost filtering
   - Cluster metrics integration
   - Anomaly filtering by payer

### Phase 6: CLI Integration (Week 3-4)
**Files:** `cli/internal/cmd/report_generate.go`, `cli/internal/report/`

1. Extend `report generate` command:
   - Add `executive-summary` template
   - Wire up exec summary data pipeline
   - Pass enriched data to template

2. Create HTML template:
   - `cli/internal/report/templates/executive-summary.html.tmpl`
   - Use Jinja2 syntax (gonja compatible)
   - Charts, tables, KPI cards

### Phase 7: Testing & Validation (Week 4)
**Files:** `core/report/execsummary/*_test.go`

1. Unit tests for each function:
   - Date utilities
   - Enrichment logic
   - Anomaly detection (with known datasets)
   - KPI calculations

2. Integration tests:
   - End-to-end with sample CSV/JSON
   - Compare outputs to Python baseline

## Key Differences: Python vs Go

### 1. Pandas DataFrame → Slices & Maps
- **Python:** `df.groupby('account_id')['cost'].sum()`
- **Go:** Manual iteration + map accumulation

### 2. Null Handling
- **Python:** `df.fillna('Unmapped')`
- **Go:** Explicit checks, pointer types for nullable fields

### 3. Time Handling
- **Python:** `date.today()`, `strftime()`
- **Go:** `time.Now()`, `time.Format()`

### 4. Statistical Functions
- **Python:** `series.mean()`, `series.std()`
- **Go:** Implement or use `gonum.org/v1/gonum/stat`

## Dependencies

### Go Libraries
- `time` (stdlib) - date/time operations
- `encoding/json`, `encoding/csv` (stdlib) - data loading
- `gonum.org/v1/gonum/stat` - statistical functions (mean, stddev)
- `github.com/nikolalohinski/gonja` - Jinja2 templating (already in use)

### External Data Sources (unchanged)
- Cost Explorer CSV/JSON export
- Account mapping CSV
- Cluster counts CSV
- Savings Plan data JSON

## Success Criteria

1. ✅ All Python analytics functions ported to Go
2. ✅ Unit tests achieving >80% coverage
3. ✅ Executive summary HTML report generated via CLI
4. ✅ Output matches Python baseline (within rounding tolerance)
5. ✅ Performance: <5s for typical dataset (12 months, 100 accounts)

## Risks & Mitigations

| Risk | Mitigation |
|------|------------|
| Pandas semantics mismatch | Write comprehensive unit tests, compare outputs |
| Statistical precision differences | Use `gonum` for consistency, test edge cases |
| Template rendering differences | Use gonja (Jinja2 compatible), test templates |
| Large dataset performance | Profile, optimize aggregations, add indexes |

## Next Steps

1. Start with Phase 1 (core analytics)
2. Create sample test data from legacy system
3. Incremental migration with continuous testing
4. Preserve Python version for validation during migration
