// env.go parses shell export output from external login tools into AWS credential fields.
package aws

import "strings"

// ParseEnvOutput parses shell-style AWS credential env lines (export KEY=value) into a map.
func ParseEnvOutput(stdout string) map[string]string {
	envMap := make(map[string]string)
	for _, line := range strings.Split(strings.TrimSpace(stdout), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "export ") {
			line = line[7:]
		}
		if idx := strings.Index(line, "="); idx >= 0 {
			key := strings.TrimSpace(line[:idx])
			val := strings.TrimSpace(line[idx+1:])
			val = strings.Trim(val, `"'`)
			envMap[key] = val
		}
	}
	return envMap
}

// SessionFromEnvMap builds a ProfileSession from parsed env vars.
func SessionFromEnvMap(env map[string]string) ProfileSession {
	region := env["AWS_REGION"]
	if region == "" {
		region = "us-east-1"
	}
	return ProfileSession{
		AccessKeyID:     env["AWS_ACCESS_KEY_ID"],
		SecretAccessKey: env["AWS_SECRET_ACCESS_KEY"],
		SessionToken:    env["AWS_SESSION_TOKEN"],
		Region:          region,
	}
}
