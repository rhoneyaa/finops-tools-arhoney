// cost_defaults.go validates cost period defaults in the finops config file.
package configstore

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/openshift-online/finops-tools/core/cost"
)

// Cost period default FQNs (defaults.<name> in config YAML).
const (
	DefaultFQNCostDays              = "cost.days"
	DefaultFQNCostMonths            = "cost.months"
	DefaultFQNCostFrom              = "cost.from"
	DefaultFQNCostTo                = "cost.to"
	DefaultFQNCostExcludeRecentDays = "cost.exclude_recent_days"
)

var costPeriodModeKeys = []string{
	DefaultFQNCostDays,
	DefaultFQNCostMonths,
	DefaultFQNCostFrom,
}

// ValidateCostPeriodDefaults reports conflicting period mode keys in config.
func (f File) ValidateCostPeriodDefaults() error {
	n := 0
	for _, key := range costPeriodModeKeys {
		if _, ok := f.Default(key); ok {
			n++
		}
	}
	if n > 1 {
		return fmt.Errorf("config has conflicting cost period defaults (set only one of %s, %s, %s)",
			DefaultFQNCostDays, DefaultFQNCostMonths, DefaultFQNCostFrom)
	}
	if v, ok := f.Default(DefaultFQNCostTo); ok {
		if _, hasFrom := f.Default(DefaultFQNCostFrom); !hasFrom {
			return fmt.Errorf("config %s requires %s", DefaultFQNCostTo, DefaultFQNCostFrom)
		}
		_ = v
	}
	return nil
}

func validateCostDefaultValue(fqn, value string) error {
	switch fqn {
	case DefaultFQNCostDays, DefaultFQNCostMonths:
		n, err := parsePositiveInt(value, fqn)
		if err != nil {
			return err
		}
		if n <= 0 {
			return fmt.Errorf("%s must be positive", fqn)
		}
	case DefaultFQNCostExcludeRecentDays:
		n, err := parseNonNegativeInt(value, fqn)
		if err != nil {
			return err
		}
		if n < 0 {
			return fmt.Errorf("%s must be >= 0", fqn)
		}
	case DefaultFQNCostFrom, DefaultFQNCostTo:
		if _, err := cost.ParseDate(value); err != nil {
			return fmt.Errorf("%s: %w", fqn, err)
		}
	default:
		return fmt.Errorf("unknown cost default %q", fqn)
	}
	return nil
}

func parsePositiveInt(value, name string) (int, error) {
	value = strings.TrimSpace(value)
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer %q", name, value)
	}
	return n, nil
}

func parseNonNegativeInt(value, name string) (int, error) {
	value = strings.TrimSpace(value)
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: invalid integer %q", name, value)
	}
	return n, nil
}
