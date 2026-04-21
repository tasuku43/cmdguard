package config

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/tasuku43/cmdproxy/internal/contract"
	"github.com/tasuku43/cmdproxy/internal/domain/policy"
	"gopkg.in/yaml.v3"
)

const LayerUser = "user"

type File struct {
	Rules []policy.RuleSpec `yaml:"rules"`
}

type Source = policy.Source

type Loaded struct {
	Rules  []policy.Rule
	Files  []Source
	Errors []error
}

type evalFile struct {
	Rules []evalRuleSpec
}

type evalRuleSpec struct {
	ID      string
	Pattern string
	Match   policy.MatchSpec
	Reject  policy.RejectSpec
	Rewrite policy.RewriteSpec
}

type evalCacheFile struct {
	Version         int              `json:"version"`
	SourcePath      string           `json:"source_path"`
	SourceHash      string           `json:"source_hash"`
	CmdproxyVersion string           `json:"cmdproxy_version,omitempty"`
	VerifiedAt      string           `json:"verified_at,omitempty"`
	CompiledRules   []evalCachedRule `json:"compiled_rules"`
}

type evalCachedRule struct {
	ID      string             `json:"id"`
	Pattern string             `json:"pattern"`
	Match   policy.MatchSpec   `json:"match,omitempty"`
	Reject  policy.RejectSpec  `json:"reject,omitempty"`
	Rewrite policy.RewriteSpec `json:"rewrite,omitempty"`
}

func ConfigPaths(home string, xdgConfigHome string) []Source {
	userConfigBase := xdgConfigHome
	if userConfigBase == "" {
		userConfigBase = filepath.Join(home, ".config")
	}
	return []Source{{
		Layer: LayerUser,
		Path:  filepath.Join(userConfigBase, "cmdproxy", "cmdproxy.yml"),
	}}
}

func HookCacheDir(home string, xdgCacheHome string) string {
	dirs := HookCacheDirs(home, xdgCacheHome)
	if len(dirs) == 0 {
		return filepath.Join(home, ".cache", "cmdproxy")
	}
	return dirs[0]
}

func HookCacheDirs(home string, xdgCacheHome string) []string {
	seen := map[string]struct{}{}
	var dirs []string
	add := func(base string) {
		if strings.TrimSpace(base) == "" {
			return
		}
		path := filepath.Join(base, "cmdproxy")
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		dirs = append(dirs, path)
	}
	add(filepath.Join(os.TempDir(), "cmdproxy-"+shortHash(home)))
	add(filepath.Join(home, ".cache"))
	add(xdgCacheHome)
	return dirs
}

func LoadEffective(home string, xdgConfigHome string) Loaded {
	return loadEffectiveWithLoader(home, xdgConfigHome, LoadFileIfPresent)
}

func LoadEffectiveForHook(home string, xdgConfigHome string, xdgCacheHome string) Loaded {
	loader := func(src Source) ([]policy.Rule, error) {
		return LoadVerifiedFileForHook(src, HookCacheDirs(home, xdgCacheHome))
	}
	return loadEffectiveWithLoader(home, xdgConfigHome, loader)
}

func loadEffectiveWithLoader(home string, xdgConfigHome string, loader func(Source) ([]policy.Rule, error)) Loaded {
	var loaded Loaded
	for _, src := range ConfigPaths(home, xdgConfigHome) {
		rules, err := loader(src)
		if err != nil {
			loaded.Errors = append(loaded.Errors, err)
			continue
		}
		if len(rules) == 0 {
			continue
		}
		loaded.Files = append(loaded.Files, src)
		loaded.Rules = append(loaded.Rules, rules...)
	}
	loaded.Errors = append(loaded.Errors, policy.ValidateDuplicateIDs(loaded.Rules)...)
	return loaded
}

func LoadFileIfPresent(src Source) ([]policy.Rule, error) {
	data, err := readConfigFile(src)
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}
	file, err := decodeFullFile(src, data)
	if err != nil {
		return nil, err
	}
	issues := validateFile(file)
	if len(issues) > 0 {
		for i := range issues {
			issues[i] = fmt.Sprintf("%s config %s: %s", src.Layer, src.Path, issues[i])
		}
		return nil, &policy.ValidationError{Issues: issues}
	}
	rules := make([]policy.Rule, 0, len(file.Rules))
	for _, spec := range file.Rules {
		rules = append(rules, policy.NewRule(spec, src))
	}
	return rules, nil
}

func LoadFileForEvalIfPresent(src Source, cacheDir string) ([]policy.Rule, error) {
	return loadFileForEval(src, cacheDir, false, "")
}

func LoadVerifiedFileForHook(src Source, cacheDirs []string) ([]policy.Rule, error) {
	return loadVerifiedFileForHook(src, cacheDirs)
}

func VerifyFile(src Source, cacheDir string, cmdproxyVersion string) ([]policy.Rule, error) {
	return compileAndWriteEvalFile(src, cacheDir, cmdproxyVersion)
}

func VerifyFileToAllCaches(src Source, cacheDirs []string, cmdproxyVersion string) ([]policy.Rule, error) {
	var rules []policy.Rule
	var errs []string
	success := false
	for i, cacheDir := range cacheDirs {
		loadedRules, err := compileAndWriteEvalFile(src, cacheDir, cmdproxyVersion)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		if !success || i == 0 {
			rules = loadedRules
		}
		success = true
	}
	if !success {
		if len(errs) == 0 {
			return nil, fmt.Errorf("failed to write verified artifacts")
		}
		return nil, errors.New(strings.Join(errs, "; "))
	}
	return rules, nil
}

func loadFileForEval(src Source, cacheDir string, requireVerified bool, cmdproxyVersion string) ([]policy.Rule, error) {
	data, err := readConfigFile(src)
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}
	sourceHash := contentHash(data)
	cachePath := cachePathForHash(cacheDir, sourceHash)
	if rules, ok := loadEvalCache(src, cachePath, sourceHash, requireVerified); ok {
		return rules, nil
	}
	if requireVerified {
		return nil, fmt.Errorf("%s config %s changed since last verify; run cmdproxy verify", src.Layer, src.Path)
	}
	return compileEvalData(src, cacheDir, cmdproxyVersion, data, sourceHash)
}

func compileAndWriteEvalFile(src Source, cacheDir string, cmdproxyVersion string) ([]policy.Rule, error) {
	data, err := readConfigFile(src)
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}
	sourceHash := contentHash(data)
	return compileEvalData(src, cacheDir, cmdproxyVersion, data, sourceHash)
}

func compileEvalData(src Source, cacheDir string, cmdproxyVersion string, data string, sourceHash string) ([]policy.Rule, error) {
	file, err := decodeEvalFile(src, data)
	if err != nil {
		return nil, err
	}
	issues := validateEvalFile(file)
	if len(issues) > 0 {
		for i := range issues {
			issues[i] = fmt.Sprintf("%s config %s: %s", src.Layer, src.Path, issues[i])
		}
		return nil, &policy.ValidationError{Issues: issues}
	}
	rules := make([]policy.Rule, 0, len(file.Rules))
	cached := make([]evalCachedRule, 0, len(file.Rules))
	for _, spec := range file.Rules {
		ruleSpec := policy.RuleSpec{
			ID:      spec.ID,
			Pattern: spec.Pattern,
			Matcher: spec.Match,
			Reject:  spec.Reject,
			Rewrite: spec.Rewrite,
		}
		rules = append(rules, policy.NewRule(ruleSpec, src))
		cached = append(cached, evalCachedRule{
			ID:      spec.ID,
			Pattern: spec.Pattern,
			Match:   spec.Match,
			Reject:  spec.Reject,
			Rewrite: spec.Rewrite,
		})
	}
	cachePath := cachePathForHash(cacheDir, sourceHash)
	if err := writeEvalCache(cachePath, evalCacheFile{
		Version:         1,
		SourcePath:      src.Path,
		SourceHash:      sourceHash,
		CmdproxyVersion: cmdproxyVersion,
		VerifiedAt:      time.Now().UTC().Format(time.RFC3339),
		CompiledRules:   cached,
	}); err != nil {
		return nil, err
	}
	pruneOldEvalCaches(cacheDir, cachePath)
	return rules, nil
}

func loadVerifiedFileForHook(src Source, cacheDirs []string) ([]policy.Rule, error) {
	data, err := readConfigFile(src)
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, nil
	}
	sourceHash := contentHash(data)
	for _, cacheDir := range cacheDirs {
		cachePath := cachePathForHash(cacheDir, sourceHash)
		if rules, ok := loadEvalCache(src, cachePath, sourceHash, true); ok {
			return rules, nil
		}
	}
	return nil, fmt.Errorf("%s config %s changed since last verify; verified artifact not found in %s; run cmdproxy verify", src.Layer, src.Path, strings.Join(cacheDirs, ", "))
}

func cachePathForHash(cacheDir string, sourceHash string) string {
	return filepath.Join(cacheDir, "compiled-rules-"+sourceHash+".json")
}

func contentHash(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

func readConfigFile(src Source) (string, error) {
	data, err := os.ReadFile(src.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", fmt.Errorf("%s config read failed: %w", src.Layer, err)
	}
	if strings.TrimSpace(string(data)) == "" {
		return "", fmt.Errorf("%s config %s is empty", src.Layer, src.Path)
	}
	return string(data), nil
}

func decodeFullFile(src Source, data string) (File, error) {
	dec := yaml.NewDecoder(strings.NewReader(data))
	dec.KnownFields(true)
	var file File
	if err := dec.Decode(&file); err != nil {
		return File{}, fmt.Errorf("%s config %s is invalid: %w", src.Layer, src.Path, err)
	}
	return file, nil
}

func decodeEvalFile(src Source, data string) (evalFile, error) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(data), &root); err != nil {
		return evalFile{}, fmt.Errorf("%s config %s is invalid: %w", src.Layer, src.Path, err)
	}
	if len(root.Content) == 0 {
		return evalFile{}, fmt.Errorf("%s config %s is invalid: empty YAML document", src.Layer, src.Path)
	}
	doc := root.Content[0]
	if doc.Kind != yaml.MappingNode {
		return evalFile{}, fmt.Errorf("%s config %s is invalid: top-level must be a mapping", src.Layer, src.Path)
	}
	file := evalFile{}
	seenTopLevel := map[string]struct{}{}
	for i := 0; i < len(doc.Content); i += 2 {
		key := doc.Content[i]
		val := doc.Content[i+1]
		if _, ok := seenTopLevel[key.Value]; ok {
			continue
		}
		seenTopLevel[key.Value] = struct{}{}
		switch key.Value {
		case "rules":
			rules, err := decodeEvalRules(src, val)
			if err != nil {
				return evalFile{}, err
			}
			file.Rules = rules
		default:
			return evalFile{}, fmt.Errorf("%s config %s is invalid: field %q not allowed", src.Layer, src.Path, key.Value)
		}
	}
	return file, nil
}

func decodeEvalRules(src Source, node *yaml.Node) ([]evalRuleSpec, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("%s config %s is invalid: rules must be a sequence", src.Layer, src.Path)
	}
	rules := make([]evalRuleSpec, 0, len(node.Content))
	for idx, item := range node.Content {
		if item.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("%s config %s is invalid: rules[%d] must be a mapping", src.Layer, src.Path, idx)
		}
		ruleSpec, err := decodeEvalRule(src, idx, item)
		if err != nil {
			return nil, err
		}
		rules = append(rules, ruleSpec)
	}
	return rules, nil
}

func decodeEvalRule(src Source, idx int, node *yaml.Node) (evalRuleSpec, error) {
	var spec evalRuleSpec
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "id":
			if val.Kind != yaml.ScalarNode {
				return evalRuleSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].id must be a string", src.Layer, src.Path, idx)
			}
			spec.ID = val.Value
		case "pattern":
			if val.Kind != yaml.ScalarNode {
				return evalRuleSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].pattern must be a string", src.Layer, src.Path, idx)
			}
			spec.Pattern = val.Value
		case "match":
			match, err := decodeEvalMatch(src, idx, val)
			if err != nil {
				return evalRuleSpec{}, err
			}
			spec.Match = match
		case "reject":
			reject, err := decodeEvalReject(src, idx, val)
			if err != nil {
				return evalRuleSpec{}, err
			}
			spec.Reject = reject
		case "rewrite":
			rewrite, err := decodeEvalRewrite(src, idx, val)
			if err != nil {
				return evalRuleSpec{}, err
			}
			spec.Rewrite = rewrite
		default:
			return evalRuleSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].%s not allowed", src.Layer, src.Path, idx, key.Value)
		}
	}
	return spec, nil
}

func decodeEvalRewrite(src Source, idx int, node *yaml.Node) (policy.RewriteSpec, error) {
	if node.Kind != yaml.MappingNode {
		return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite must be a mapping", src.Layer, src.Path, idx)
	}
	var rewrite policy.RewriteSpec
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "continue":
			if val.Kind != yaml.ScalarNode {
				return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.continue must be a boolean", src.Layer, src.Path, idx)
			}
			var enabled bool
			if err := val.Decode(&enabled); err != nil {
				return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.continue must be a boolean", src.Layer, src.Path, idx)
			}
			rewrite.Continue = enabled
		case "strict":
			if val.Kind != yaml.ScalarNode {
				return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.strict must be a boolean", src.Layer, src.Path, idx)
			}
			var enabled bool
			if err := val.Decode(&enabled); err != nil {
				return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.strict must be a boolean", src.Layer, src.Path, idx)
			}
			rewrite.Strict = &enabled
		case "unwrap_shell_dash_c":
			if val.Kind != yaml.ScalarNode {
				return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.unwrap_shell_dash_c must be a boolean", src.Layer, src.Path, idx)
			}
			var enabled bool
			if err := val.Decode(&enabled); err != nil {
				return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.unwrap_shell_dash_c must be a boolean", src.Layer, src.Path, idx)
			}
			rewrite.UnwrapShellDashC = enabled
		case "move_flag_to_env":
			spec, err := decodeEvalMoveFlagToEnv(src, idx, val)
			if err != nil {
				return policy.RewriteSpec{}, err
			}
			rewrite.MoveFlagToEnv = spec
		case "move_env_to_flag":
			spec, err := decodeEvalMoveEnvToFlag(src, idx, val)
			if err != nil {
				return policy.RewriteSpec{}, err
			}
			rewrite.MoveEnvToFlag = spec
		case "strip_command_path":
			if val.Kind != yaml.ScalarNode {
				return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.strip_command_path must be a boolean", src.Layer, src.Path, idx)
			}
			var enabled bool
			if err := val.Decode(&enabled); err != nil {
				return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.strip_command_path must be a boolean", src.Layer, src.Path, idx)
			}
			rewrite.StripCommandPath = enabled
		case "test":
			spec, err := decodeEvalRewriteTest(src, idx, val)
			if err != nil {
				return policy.RewriteSpec{}, err
			}
			rewrite.Test = spec
		case "unwrap_wrapper":
			spec, err := decodeEvalUnwrapWrapper(src, idx, val)
			if err != nil {
				return policy.RewriteSpec{}, err
			}
			rewrite.UnwrapWrapper = spec
		default:
			return policy.RewriteSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.%s not allowed", src.Layer, src.Path, idx, key.Value)
		}
	}
	return rewrite, nil
}

func decodeEvalMoveFlagToEnv(src Source, idx int, node *yaml.Node) (policy.MoveFlagToEnvSpec, error) {
	if node.Kind != yaml.MappingNode {
		return policy.MoveFlagToEnvSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.move_flag_to_env must be a mapping", src.Layer, src.Path, idx)
	}
	var spec policy.MoveFlagToEnvSpec
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "flag":
			if val.Kind != yaml.ScalarNode {
				return policy.MoveFlagToEnvSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.move_flag_to_env.flag must be a string", src.Layer, src.Path, idx)
			}
			spec.Flag = val.Value
		case "env":
			if val.Kind != yaml.ScalarNode {
				return policy.MoveFlagToEnvSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.move_flag_to_env.env must be a string", src.Layer, src.Path, idx)
			}
			spec.Env = val.Value
		default:
			return policy.MoveFlagToEnvSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.move_flag_to_env.%s not allowed", src.Layer, src.Path, idx, key.Value)
		}
	}
	return spec, nil
}

func decodeEvalMoveEnvToFlag(src Source, idx int, node *yaml.Node) (policy.MoveEnvToFlagSpec, error) {
	if node.Kind != yaml.MappingNode {
		return policy.MoveEnvToFlagSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.move_env_to_flag must be a mapping", src.Layer, src.Path, idx)
	}
	var spec policy.MoveEnvToFlagSpec
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "env":
			if val.Kind != yaml.ScalarNode {
				return policy.MoveEnvToFlagSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.move_env_to_flag.env must be a string", src.Layer, src.Path, idx)
			}
			spec.Env = val.Value
		case "flag":
			if val.Kind != yaml.ScalarNode {
				return policy.MoveEnvToFlagSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.move_env_to_flag.flag must be a string", src.Layer, src.Path, idx)
			}
			spec.Flag = val.Value
		default:
			return policy.MoveEnvToFlagSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.move_env_to_flag.%s not allowed", src.Layer, src.Path, idx, key.Value)
		}
	}
	return spec, nil
}

func decodeEvalUnwrapWrapper(src Source, idx int, node *yaml.Node) (policy.UnwrapWrapperSpec, error) {
	if node.Kind != yaml.MappingNode {
		return policy.UnwrapWrapperSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.unwrap_wrapper must be a mapping", src.Layer, src.Path, idx)
	}
	var spec policy.UnwrapWrapperSpec
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "wrappers":
			values, err := decodeStringSequence(src, idx, "rewrite.unwrap_wrapper.wrappers", val)
			if err != nil {
				return policy.UnwrapWrapperSpec{}, err
			}
			spec.Wrappers = values
		default:
			return policy.UnwrapWrapperSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.unwrap_wrapper.%s not allowed", src.Layer, src.Path, idx, key.Value)
		}
	}
	return spec, nil
}

func decodeEvalReject(src Source, idx int, node *yaml.Node) (policy.RejectSpec, error) {
	if node.Kind != yaml.MappingNode {
		return policy.RejectSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].reject must be a mapping", src.Layer, src.Path, idx)
	}
	var reject policy.RejectSpec
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "message":
			if val.Kind != yaml.ScalarNode {
				return policy.RejectSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].reject.message must be a string", src.Layer, src.Path, idx)
			}
			reject.Message = val.Value
		case "test":
			spec, err := decodeEvalRejectTest(src, idx, val)
			if err != nil {
				return policy.RejectSpec{}, err
			}
			reject.Test = spec
		default:
			return policy.RejectSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].reject.%s not allowed", src.Layer, src.Path, idx, key.Value)
		}
	}
	return reject, nil
}

func decodeEvalRejectTest(src Source, idx int, node *yaml.Node) (policy.RejectTestSpec, error) {
	if node.Kind != yaml.MappingNode {
		return policy.RejectTestSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].reject.test must be a mapping", src.Layer, src.Path, idx)
	}
	var test policy.RejectTestSpec
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "expect":
			values, err := decodeStringSequence(src, idx, "reject.test.expect", val)
			if err != nil {
				return policy.RejectTestSpec{}, err
			}
			test.Expect = values
		case "pass":
			values, err := decodeStringSequence(src, idx, "reject.test.pass", val)
			if err != nil {
				return policy.RejectTestSpec{}, err
			}
			test.Pass = values
		default:
			return policy.RejectTestSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].reject.test.%s not allowed", src.Layer, src.Path, idx, key.Value)
		}
	}
	return test, nil
}

func decodeEvalRewriteTest(src Source, idx int, node *yaml.Node) (policy.RewriteTestSpec, error) {
	if node.Kind != yaml.MappingNode {
		return policy.RewriteTestSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.test must be a mapping", src.Layer, src.Path, idx)
	}
	var test policy.RewriteTestSpec
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "expect":
			cases, err := decodeEvalRewriteExpectCases(src, idx, val)
			if err != nil {
				return policy.RewriteTestSpec{}, err
			}
			test.Expect = cases
		case "pass":
			values, err := decodeStringSequence(src, idx, "rewrite.test.pass", val)
			if err != nil {
				return policy.RewriteTestSpec{}, err
			}
			test.Pass = values
		default:
			return policy.RewriteTestSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.test.%s not allowed", src.Layer, src.Path, idx, key.Value)
		}
	}
	return test, nil
}

func decodeEvalRewriteExpectCases(src Source, idx int, node *yaml.Node) ([]policy.RewriteExpectCase, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.test.expect must be a sequence", src.Layer, src.Path, idx)
	}
	cases := make([]policy.RewriteExpectCase, 0, len(node.Content))
	for caseIdx, item := range node.Content {
		if item.Kind != yaml.MappingNode {
			return nil, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.test.expect[%d] must be a mapping", src.Layer, src.Path, idx, caseIdx)
		}
		var c policy.RewriteExpectCase
		for i := 0; i < len(item.Content); i += 2 {
			key := item.Content[i]
			val := item.Content[i+1]
			switch key.Value {
			case "in":
				if val.Kind != yaml.ScalarNode {
					return nil, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.test.expect[%d].in must be a string", src.Layer, src.Path, idx, caseIdx)
				}
				c.In = val.Value
			case "out":
				if val.Kind != yaml.ScalarNode {
					return nil, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.test.expect[%d].out must be a string", src.Layer, src.Path, idx, caseIdx)
				}
				c.Out = val.Value
			default:
				return nil, fmt.Errorf("%s config %s is invalid: rules[%d].rewrite.test.expect[%d].%s not allowed", src.Layer, src.Path, idx, caseIdx, key.Value)
			}
		}
		cases = append(cases, c)
	}
	return cases, nil
}

func decodeEvalMatch(src Source, idx int, node *yaml.Node) (policy.MatchSpec, error) {
	if node.Kind != yaml.MappingNode {
		return policy.MatchSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].match must be a mapping", src.Layer, src.Path, idx)
	}
	var match policy.MatchSpec
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i]
		val := node.Content[i+1]
		switch key.Value {
		case "command":
			if val.Kind != yaml.ScalarNode {
				return policy.MatchSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].match.command must be a string", src.Layer, src.Path, idx)
			}
			match.Command = val.Value
		case "command_in":
			values, err := decodeStringSequence(src, idx, "match.command_in", val)
			if err != nil {
				return policy.MatchSpec{}, err
			}
			match.CommandIn = values
		case "command_is_absolute_path":
			if val.Kind != yaml.ScalarNode {
				return policy.MatchSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].match.command_is_absolute_path must be a boolean", src.Layer, src.Path, idx)
			}
			var enabled bool
			if err := val.Decode(&enabled); err != nil {
				return policy.MatchSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].match.command_is_absolute_path must be a boolean", src.Layer, src.Path, idx)
			}
			match.CommandIsAbsolutePath = enabled
		case "subcommand":
			if val.Kind != yaml.ScalarNode {
				return policy.MatchSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].match.subcommand must be a string", src.Layer, src.Path, idx)
			}
			match.Subcommand = val.Value
		case "args_contains":
			values, err := decodeStringSequence(src, idx, "match.args_contains", val)
			if err != nil {
				return policy.MatchSpec{}, err
			}
			match.ArgsContains = values
		case "args_prefixes":
			values, err := decodeStringSequence(src, idx, "match.args_prefixes", val)
			if err != nil {
				return policy.MatchSpec{}, err
			}
			match.ArgsPrefixes = values
		case "env_requires":
			values, err := decodeStringSequence(src, idx, "match.env_requires", val)
			if err != nil {
				return policy.MatchSpec{}, err
			}
			match.EnvRequires = values
		case "env_missing":
			values, err := decodeStringSequence(src, idx, "match.env_missing", val)
			if err != nil {
				return policy.MatchSpec{}, err
			}
			match.EnvMissing = values
		default:
			return policy.MatchSpec{}, fmt.Errorf("%s config %s is invalid: rules[%d].match.%s not allowed", src.Layer, src.Path, idx, key.Value)
		}
	}
	return match, nil
}

func decodeStringSequence(src Source, idx int, field string, node *yaml.Node) ([]string, error) {
	if node.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("%s config %s is invalid: rules[%d].%s must be a sequence", src.Layer, src.Path, idx, field)
	}
	values := make([]string, 0, len(node.Content))
	for _, item := range node.Content {
		if item.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("%s config %s is invalid: rules[%d].%s must contain only strings", src.Layer, src.Path, idx, field)
		}
		values = append(values, item.Value)
	}
	return values, nil
}

func validateFile(file File) []string {
	var issues []string
	if len(file.Rules) == 0 {
		issues = append(issues, "rules must be non-empty")
	}
	issues = append(issues, policy.ValidateRules(file.Rules)...)
	issues = append(issues, contract.ValidateRules(file.Rules)...)
	return issues
}

func validateEvalFile(file evalFile) []string {
	var issues []string
	if len(file.Rules) == 0 {
		issues = append(issues, "rules must be non-empty")
	}
	seen := map[string]struct{}{}
	for i, r := range file.Rules {
		prefix := fmt.Sprintf("rules[%d]", i)
		if !regexp.MustCompile(`^[a-z0-9][a-z0-9-]*$`).MatchString(r.ID) {
			issues = append(issues, prefix+".id must match [a-z0-9][a-z0-9-]*")
		}
		if _, ok := seen[r.ID]; ok && r.ID != "" {
			issues = append(issues, prefix+".id duplicates another rule in the same file")
		}
		seen[r.ID] = struct{}{}
		issues = append(issues, policy.ValidateRuleMatcher(prefix, r.Pattern, r.Match)...)
		issues = append(issues, policy.ValidateDirective(prefix, r.Reject, r.Rewrite)...)
	}
	rules := make([]policy.RuleSpec, 0, len(file.Rules))
	for _, rule := range file.Rules {
		rules = append(rules, policy.RuleSpec{
			ID:      rule.ID,
			Pattern: rule.Pattern,
			Matcher: rule.Match,
			Reject:  rule.Reject,
			Rewrite: rule.Rewrite,
		})
	}
	issues = append(issues, contract.ValidateRules(rules)...)
	return issues
}

func loadEvalCache(src Source, cachePath string, sourceHash string, requireVerified bool) ([]policy.Rule, bool) {
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, false
	}
	var cache evalCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, false
	}
	if cache.Version != 1 || cache.SourcePath != src.Path || cache.SourceHash != sourceHash {
		return nil, false
	}
	if requireVerified && (strings.TrimSpace(cache.CmdproxyVersion) == "" || strings.TrimSpace(cache.VerifiedAt) == "") {
		return nil, false
	}
	rules := make([]policy.Rule, 0, len(cache.CompiledRules))
	for _, spec := range cache.CompiledRules {
		if strings.TrimSpace(spec.Pattern) != "" {
			if _, err := regexp.Compile(spec.Pattern); err != nil {
				return nil, false
			}
		}
		rules = append(rules, policy.NewRule(policy.RuleSpec{
			ID:      spec.ID,
			Pattern: spec.Pattern,
			Matcher: spec.Match,
			Reject:  spec.Reject,
			Rewrite: spec.Rewrite,
		}, src))
	}
	return rules, true
}

func writeEvalCache(cachePath string, cache evalCacheFile) error {
	if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
		return err
	}
	data, err := json.Marshal(cache)
	if err != nil {
		return err
	}
	return os.WriteFile(cachePath, data, 0o644)
}

func pruneOldEvalCaches(cacheDir string, keepPath string) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "compiled-rules-") || !strings.HasSuffix(name, ".json") {
			continue
		}
		path := filepath.Join(cacheDir, name)
		if path == keepPath {
			continue
		}
		_ = os.Remove(path)
	}
}

func shortHash(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])[:12]
}
