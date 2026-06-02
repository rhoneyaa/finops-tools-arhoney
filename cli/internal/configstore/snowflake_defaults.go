// snowflake_defaults.go defines fully qualified Snowflake configuration defaults.
package configstore

import (
	"fmt"
	"strings"
)

const (
	// DefaultFQNSnowflakeSSOIssuer selects Red Hat SSO: prod or stage.
	DefaultFQNSnowflakeSSOIssuer = "snowflake.sso_issuer"
	// DefaultFQNSnowflakeOAuthAudience is the required JWT audience for Dataverse Snowflake.
	DefaultFQNSnowflakeOAuthAudience = "snowflake.oauth_audience"
	// DefaultFQNSnowflakeAccountAlias is the default registered Snowflake account alias for snowflake commands.
	DefaultFQNSnowflakeAccountAlias = "snowflake.account_alias"
	// DefaultFQNSnowflakeWarehouse is the default warehouse when an alias omits warehouse.
	DefaultFQNSnowflakeWarehouse = "snowflake.warehouse"
	// DefaultFQNSnowflakeRole is the default role when an alias omits role.
	DefaultFQNSnowflakeRole = "snowflake.role"
	// DefaultFQNSnowflakeDatabase is the default database when an alias omits database.
	DefaultFQNSnowflakeDatabase = "snowflake.database"
	// DefaultFQNSnowflakeSchema is the default schema when an alias omits schema.
	DefaultFQNSnowflakeSchema = "snowflake.schema"
)

const (
	defaultSnowflakeOAuthAudience = "dataverse-snowflake"
	defaultSnowflakeSSOIssuer     = "prod"
)

func validateSnowflakeDefaultValue(fqn, value string) error {
	switch fqn {
	case DefaultFQNSnowflakeSSOIssuer:
		switch strings.ToLower(strings.TrimSpace(value)) {
		case "", "prod", "production", "stage", "staging":
			return nil
		default:
			return fmt.Errorf("snowflake.sso_issuer must be prod or stage")
		}
	case DefaultFQNSnowflakeOAuthAudience:
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("snowflake.oauth_audience cannot be empty")
		}
		return nil
	case DefaultFQNSnowflakeAccountAlias:
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("snowflake.account_alias cannot be empty")
		}
		return nil
	case DefaultFQNSnowflakeWarehouse, DefaultFQNSnowflakeRole, DefaultFQNSnowflakeDatabase, DefaultFQNSnowflakeSchema:
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("%s cannot be empty", fqn)
		}
		return nil
	default:
		return fmt.Errorf("unknown snowflake default %q", fqn)
	}
}

// SnowflakeSSOIssuer returns the configured Red Hat SSO environment (prod or stage).
func (f File) SnowflakeSSOIssuer() string {
	if v, ok := f.Default(DefaultFQNSnowflakeSSOIssuer); ok && strings.TrimSpace(v) != "" {
		return strings.ToLower(strings.TrimSpace(v))
	}
	return defaultSnowflakeSSOIssuer
}

// SnowflakeOAuthAudience returns the JWT audience required by Dataverse Snowflake.
func (f File) SnowflakeOAuthAudience() string {
	if v, ok := f.Default(DefaultFQNSnowflakeOAuthAudience); ok && strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return defaultSnowflakeOAuthAudience
}
