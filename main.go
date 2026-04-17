package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/kitproj/coding-context-cli/pkg/codingcontext"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/selectors"
	"github.com/kitproj/coding-context-cli/pkg/codingcontext/taskparser"
)

var (
	errInvalidUsage          = errors.New("invalid usage: expected one task name argument and optional user-prompt")
	errWriteRulesNoAgent     = errors.New("-w flag requires an agent to be specified (via task 'agent' field or -a flag)")
	errNoUserRulePath        = errors.New("no user rule path available for agent")
	errRulesPathEscapesHome  = errors.New("rules path escapes home directory")
	errAgentFlagsMutExcl     = errors.New("-a and -A flags are mutually exclusive")
)

type cliConfig struct {
	workDir            string
	resume             bool
	skipBootstrap      bool
	writeRules         bool
	agent              codingcontext.Agent
	lenientAgent       codingcontext.Agent
	params             taskparser.Params
	includes           selectors.Selectors
	searchPaths        []string
	lenientSearchPaths []string
	manifestURL        string
	taskName           string
	userPrompt         string
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	if err := run(ctx, logger); err != nil {
		logger.Error("Error", "error", err)
		cancel()
		os.Exit(1)
	}

	cancel()
}

func run(ctx context.Context, logger *slog.Logger) error {
	cfg, err := parseFlags(logger)
	if err != nil {
		return err
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	cfg.searchPaths = append(cfg.searchPaths, "file://"+cfg.workDir)
	cfg.searchPaths = append(cfg.searchPaths, "file://"+homeDir)

	cc := codingcontext.New(
		codingcontext.WithParams(cfg.params),
		codingcontext.WithSelectors(cfg.includes),
		codingcontext.WithSearchPaths(cfg.searchPaths...),
		codingcontext.WithLenientSearchPaths(cfg.lenientSearchPaths...),
		codingcontext.WithLogger(logger),
		codingcontext.WithResume(cfg.resume),
		codingcontext.WithBootstrap(!cfg.skipBootstrap),
		codingcontext.WithManifestURL(cfg.manifestURL),
		codingcontext.WithUserPrompt(cfg.userPrompt),
		codingcontext.WithAgent(cfg.agent),
		codingcontext.WithLenientAgent(cfg.lenientAgent),
	)

	result, err := cc.Run(ctx, cfg.taskName)
	if err != nil {
		flag.Usage()

		return fmt.Errorf("%w", err)
	}

	outputContent, err := buildOutputContent(result, cfg, homeDir, logger)
	if err != nil {
		return err
	}

	if _, err := os.Stdout.Write(append([]byte(outputContent), '\n')); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}

func buildOutputContent(
	result *codingcontext.Result, cfg *cliConfig, homeDir string, logger *slog.Logger,
) (string, error) {
	if !cfg.writeRules {
		return result.Prompt, nil
	}

	if err := writeRulesToAgent(result, homeDir, cfg.skipBootstrap, logger); err != nil {
		return "", err
	}

	return result.Task.Content, nil
}

func parseFlags(logger *slog.Logger) (*cliConfig, error) {
	cfg := &cliConfig{
		params:   make(taskparser.Params),
		includes: make(selectors.Selectors),
	}

	flag.StringVar(&cfg.workDir, "C", ".", "Change to directory before doing anything.")
	flag.BoolVar(&cfg.resume, "r", false,
		"Resume mode: set 'resume=true' selector to filter tasks by their frontmatter resume field.")
	flag.BoolVar(&cfg.skipBootstrap, "skip-bootstrap", false,
		"Skip bootstrap: skip discovering rules, skills, and running bootstrap scripts.")
	flag.BoolVar(&cfg.writeRules, "w", false,
		"Write rules to the agent's user rules path and only print the prompt to stdout. "+
			"Requires agent (via task 'agent' field or -a flag).")
	flag.Var(&cfg.agent, "a",
		"Target agent to use. Required when using -w to write rules to the agent's user rules path. "+
			"Supported agents: cursor, opencode, copilot, claude, gemini, augment, windsurf, codex.")
	flag.Var(&cfg.lenientAgent, "A",
		"Target agent with lenient error handling (errors are warnings, missing skill names inferred from directory). "+
			"Mutually exclusive with -a. Supported agents: cursor, opencode, copilot, claude, gemini, augment, windsurf, codex.")
	flag.Var(&cfg.params, "p", "Parameter to substitute in the prompt. Can be specified multiple times as key=value.")
	flag.Var(&cfg.includes, "s", "Include rules with matching frontmatter. Can be specified multiple times as key=value.")
	flag.Func("d",
		"Directory containing rules and tasks (strict: errors are fatal). Can be specified multiple times. "+
			"Supports various protocols via go-getter (http://, https://, git::, s3::, file:// etc.).",
		func(s string) error {
			cfg.searchPaths = append(cfg.searchPaths, s)

			return nil
		})
	flag.Func("D",
		"Directory containing rules and tasks (lenient: errors are warnings). Can be specified multiple times. "+
			"Supports various protocols via go-getter (http://, https://, git::, s3::, file:// etc.).",
		func(s string) error {
			cfg.lenientSearchPaths = append(cfg.lenientSearchPaths, s)

			return nil
		})
	flag.StringVar(&cfg.manifestURL, "m", "",
		"Go Getter URL to a manifest file containing search paths (one per line). Every line is included as-is.")

	setupUsage(logger)
	flag.Parse()

	return parseFlagArgs(cfg)
}

func setupUsage(logger *slog.Logger) {
	flag.Usage = func() {
		logger.Info("Usage:")
		logger.Info("  coding-context [options] <task-name> [user-prompt]")
		logger.Info("")
		logger.Info("The task-name is the name of a task file to look up in task search paths (.agents/tasks).")
		logger.Info("The user-prompt is optional text to append to the task. It can contain slash commands")
		logger.Info("(e.g., '/command-name') which will be expanded, and parameter substitution (${param}).")
		logger.Info("")
		logger.Info("Task content can contain slash commands (e.g., '/command-name arg') which reference")
		logger.Info("command files in command search paths (.cursor/commands, .agents/commands, etc.).")
		logger.Info("")
		logger.Info("Options:")
		flag.PrintDefaults()
	}
}

func parseFlagArgs(cfg *cliConfig) (*cliConfig, error) {
	if cfg.agent.IsSet() && cfg.lenientAgent.IsSet() {
		return nil, errAgentFlagsMutExcl
	}

	args := flag.Args()

	const maxArgs = 2

	if len(args) < 1 || len(args) > maxArgs {
		flag.Usage()

		return nil, errInvalidUsage
	}

	cfg.taskName = args[0]

	if len(args) == maxArgs {
		cfg.userPrompt = args[1]
	}

	return cfg, nil
}

func writeRulesToAgent(result *codingcontext.Result, homeDir string, skipBootstrap bool, logger *slog.Logger) error {
	if !result.Agent.IsSet() {
		return errWriteRulesNoAgent
	}

	if skipBootstrap {
		return nil
	}

	relativePath := result.Agent.UserRulePath()
	if relativePath == "" {
		return errNoUserRulePath
	}

	rulesFile := filepath.Join(homeDir, relativePath)
	rulesFile = filepath.Clean(rulesFile)

	homeDirAbs := filepath.Clean(homeDir) + string(filepath.Separator)
	if !strings.HasPrefix(rulesFile, homeDirAbs) && rulesFile != filepath.Clean(homeDir) {
		return fmt.Errorf("%w: %s", errRulesPathEscapesHome, rulesFile)
	}

	rulesDir := filepath.Dir(rulesFile)

	const dirMode = 0o750

	// #nosec G703 -- rulesDir is validated to be within homeDir via rulesFile check above
	if err := os.MkdirAll(rulesDir, dirMode); err != nil {
		return fmt.Errorf("failed to create rules directory %s: %w", rulesDir, err)
	}

	var rulesContent strings.Builder

	for i, rule := range result.Rules {
		if i > 0 {
			rulesContent.WriteString("\n\n")
		}

		rulesContent.WriteString(strings.TrimSpace(rule.Content))
	}

	rulesContent.WriteString("\n")

	const fileMode = 0o600

	// #nosec G703 -- rulesFile is validated to be within homeDir via HasPrefix check above
	if err := os.WriteFile(rulesFile, []byte(rulesContent.String()), fileMode); err != nil {
		return fmt.Errorf("failed to write rules to %s: %w", rulesFile, err)
	}

	logger.Info("Rules written", "path", rulesFile)

	return nil
}
