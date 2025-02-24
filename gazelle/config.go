package gazelle

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/bazelbuild/bazel-gazelle/config"
	"github.com/bazelbuild/bazel-gazelle/rule"
	"github.com/cgrindel/swift_bazel/gazelle/internal/reslog"
	"github.com/cgrindel/swift_bazel/gazelle/internal/swiftbin"
	"github.com/cgrindel/swift_bazel/gazelle/internal/swiftcfg"
)

// Register Flags

func (*swiftLang) RegisterFlags(fs *flag.FlagSet, cmd string, c *config.Config) {
	// Initialize location for custom configuration
	sc := swiftcfg.NewSwiftConfig()

	fs.StringVar(
		&sc.DependencyIndexRel,
		"swift_dependency_index",
		swiftcfg.DefaultDependencyIndexBasename,
		"the location of the dependency index JSON file",
	)

	switch cmd {
	case "fix", "update":
		fs.StringVar(
			&sc.ResolutionLogPath,
			"resolution_log",
			"",
			"the location of the resolution log file",
		)
	case "update-repos":
		fs.BoolVar(
			&sc.UpdatePkgsToLatest,
			"swift_update_packages_to_latest",
			false,
			"determines whether to update the Swift packages to their latest eligible version.")
		fs.BoolVar(
			&sc.UpdateBzlmodUseRepoNames,
			"update_bzlmod_use_repo_names",
			false,
			"determines whether to update the use_repo names in your MODULE.bazel file with the appropriate stanzas.")
		fs.BoolVar(
			&sc.PrintBzlmodStanzas,
			"print_bzlmod_stanzas",
			false,
			"determines whether to print the bzlmod stanzas to stdout.")
		fs.BoolVar(
			&sc.UpdateBzlmodStanzas,
			"update_bzlmod_stanzas",
			false,
			"determines whether to update your MODULE.bazel file with the appropriate stanzas.")
		fs.StringVar(
			&sc.BazelModuleRel,
			"bazel_module",
			"MODULE.bazel",
			"the location of the MODULE.bazel file")
		fs.BoolVar(
			&sc.GenerateSwiftDepsForWorkspace,
			"generate_swift_deps_for_workspace",
			false,
			"determines whether to generate swift deps for workspace (e.g. swift_deps.bzl).")
	}

	// Store the config for later steps
	swiftcfg.SetSwiftConfig(c, sc)
}

func (sl *swiftLang) CheckFlags(fs *flag.FlagSet, c *config.Config) error {
	var err error
	sc := swiftcfg.GetSwiftConfig(c)

	// GH021: Add flag so that the client can tell us which Swift to use.

	if sc.ResolutionLogPath != "" {
		sc.ResolutionLogFile, err = os.Create(sc.ResolutionLogPath)
		if err != nil {
			return err
		}
		sc.ResolutionLogger = reslog.NewLoggerFromWriter(sc.ResolutionLogFile)
	}

	// Find the Swift executable
	if sc.SwiftBinPath, err = swiftbin.FindSwiftBinPath(); err != nil {
		return err
	}

	// Initialize the module index path. We cannot initialize this path until we get into
	// CheckFlags.
	if sc.DependencyIndexPath == "" {
		sc.DependencyIndexPath = filepath.Join(c.RepoRoot, sc.DependencyIndexRel)
	}

	if sc.BazelModulePath == "" {
		sc.BazelModulePath = filepath.Join(c.RepoRoot, sc.BazelModuleRel)
	}

	// Attempt to load the module index. This is created by update-repos if the client is using
	// external Swift packages (e.g. swift_pacakge).
	if err = sc.LoadDependencyIndex(); err != nil {
		return err
	}
	// Index any of repository rules (e.g. http_archive) that may contain Swift targets.
	for _, r := range c.Repos {
		if err := sc.DependencyIndex.IndexRepoRule(r, c.RepoRoot); err != nil {
			return err
		}
	}

	return nil
}

// Directives

const defaultModuleNameDirective = "swift_default_module_name"

func (*swiftLang) KnownDirectives() []string {
	return []string{defaultModuleNameDirective}
}

func (*swiftLang) Configure(c *config.Config, rel string, f *rule.File) {
	if f == nil {
		return
	}
	sc := swiftcfg.GetSwiftConfig(c)
	for _, d := range f.Directives {
		switch d.Key {
		case defaultModuleNameDirective:
			sc.DefaultModuleNames[rel] = d.Value
		}
	}
}
