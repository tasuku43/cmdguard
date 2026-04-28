package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/tasuku43/cc-bash-guard/internal/infra"
)

func RunInit(env Env) (InitResult, error) {
	return RunInitWithOptions(env, InitOptions{})
}

func RunInitWithOptions(env Env, opts InitOptions) (InitResult, error) {
	configDir := filepath.Join(userConfigBase(env.Home, env.XDGConfigHome), "cc-bash-guard")
	if err := infra.MkdirAll(configDir, 0o755); err != nil {
		return InitResult{}, err
	}

	config := starterConfig
	if opts.Profile != "" {
		profile, ok := LookupInitProfile(opts.Profile)
		if !ok {
			return InitResult{}, fmt.Errorf("unknown profile %q. Supported profiles: %s", opts.Profile, strings.Join(InitProfileNames(), ", "))
		}
		config = profile.Config
	}

	configPath := filepath.Join(configDir, "cc-bash-guard.yml")
	created := false
	exists, err := infra.Exists(configPath)
	if err != nil {
		return InitResult{}, err
	}
	if !exists {
		if err := infra.WriteFile(configPath, []byte(config), 0o644); err != nil {
			return InitResult{}, err
		}
		created = true
	}

	claudeSettings := filepath.Join(env.Home, ".claude", "settings.json")
	settingsDetected, err := infra.Exists(claudeSettings)
	if err != nil && !errors.Is(err, filepath.ErrBadPattern) {
		return InitResult{}, err
	}

	return InitResult{
		ConfigPath:             configPath,
		Created:                created,
		Profile:                opts.Profile,
		ClaudeSettingsPath:     claudeSettings,
		ClaudeSettingsDetected: settingsDetected,
		HookSnippet:            `{"matcher":"Bash","hooks":[{"type":"command","command":"cc-bash-guard hook"}]}`,
	}, nil
}

func userConfigBase(home string, xdgConfigHome string) string {
	if xdgConfigHome != "" {
		return xdgConfigHome
	}
	return filepath.Join(home, ".config")
}

const starterConfig = `permission:
  deny:
    - patterns:
        - "^git\\s+-C\\b"
      message: "git -C is blocked. Change into the target directory and rerun the command."
      test:
        deny:
          - "git -C repos/foo status"
        abstain:
          - "git status"
test:
  - in: "git -C repos/foo status"
    decision: deny
`
