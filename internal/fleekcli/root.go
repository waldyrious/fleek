package fleekcli

import (
	"io"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/spf13/cobra"
	"github.com/ublue-os/fleek/fin"
	"github.com/ublue-os/fleek/internal/fleek"
	"github.com/ublue-os/fleek/internal/fleekcli/usererr"
)

var cfg *fleek.Config
var cfgFound bool

type rootCmdFlags struct {
	quiet   bool
	verbose bool
}

func RootCmd() *cobra.Command {
	flags := rootCmdFlags{}
	command := &cobra.Command{
		Use:   app.Trans("fleek.use"),
		Short: app.Trans("fleek.short"),
		Long:  app.Trans("fleek.long"),

		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if flags.quiet {
				cmd.SetErr(io.Discard)
			}
			fin.Debug.Println("debug enabled")
			info, ok := debug.ReadBuildInfo()
			if ok {

				fin.Debug.Println(info.String())

			}
			warn := os.Getenv("WARN_FLEEK")
			if warn == "" {
				ex, err := os.Executable()
				if err != nil {
					panic(err)
				}
				exePath := filepath.Dir(ex)
				fin.Debug.Println("installed at: " + exePath)
				// this is pretty hokey, but it's the best we can do for now
				// /nix/var/nix is the actual store, but macos reports the binary
				// location as the symlinked path instead of actual store path
				if !strings.Contains(exePath, "nix") {
					fin.Warning.Println(app.Trans("fleek.unsupported"))
				}
			}

			// try to get the config, which may not exist yet
			c, err := fleek.ReadConfig("")
			if err == nil {
				if flags.verbose {
					fin.Info.Println(app.Trans("fleek.configLoaded"))
				}
				cfg = c
				cfgFound = true
			} else {
				cfg = &fleek.Config{}
				cfgFound = false
			}
			if cfg != nil {
				cfg.Quiet = flags.quiet
				cfg.Verbose = flags.verbose
				fin.Debug.Printfln("git autopush: %v", cfg.Git.AutoPush)
				fin.Debug.Printfln("git autocommit: %v", cfg.Git.AutoCommit)
				fin.Debug.Printfln("git autopull: %v", cfg.Git.AutoPull)
				if cfg.Ejected {
					if cmd.Name() != app.Trans("apply.use") {
						fin.Error.Println(app.Trans("eject.ejected"))
						os.Exit(1)
					}
				}

			}

		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if flags.quiet {
				cmd.SetErr(io.Discard)
			}
			fin.Debug.Printfln("git autopush: %v", cfg.Git.AutoPush)
			fin.Debug.Printfln("git autocommit: %v", cfg.Git.AutoCommit)
			fin.Debug.Printfln("git autopull: %v", cfg.Git.AutoPull)

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	initGroup := &cobra.Group{
		ID:    "init",
		Title: app.Trans("global.initGroup"),
	}
	fleekGroup := &cobra.Group{
		ID:    "fleek",
		Title: app.Trans("global.fleekGroup"),
	}

	packageGroup := &cobra.Group{
		ID:    "package",
		Title: app.Trans("global.packageGroup"),
	}

	command.AddGroup(initGroup, packageGroup, fleekGroup)
	addCmd := AddCommand()
	addCmd.GroupID = packageGroup.ID

	removeCmd := RemoveCommand()
	removeCmd.GroupID = packageGroup.ID

	showCmd := ShowCmd()
	showCmd.GroupID = fleekGroup.ID

	applyCmd := ApplyCommand()
	applyCmd.GroupID = fleekGroup.ID

	updateCmd := UpdateCommand()
	updateCmd.GroupID = packageGroup.ID

	initCmd := InitCommand()
	initCmd.GroupID = initGroup.ID
	joinCmd := JoinCommand()
	joinCmd.GroupID = initGroup.ID
	ejectCmd := EjectCommand()
	ejectCmd.GroupID = fleekGroup.ID
	generateCmd := GenerateCommand()
	generateCmd.GroupID = fleekGroup.ID
	searchCmd := SearchCommand()
	searchCmd.GroupID = packageGroup.ID

	infoCmd := InfoCommand()
	infoCmd.GroupID = packageGroup.ID
	manCmd := ManCommand()

	docsCmd := genDocsCmd()
	command.AddCommand(docsCmd)
	command.AddCommand(manCmd)
	command.AddCommand(showCmd)

	command.AddCommand(addCmd)
	command.AddCommand(removeCmd)
	command.AddCommand(applyCmd)
	command.AddCommand(updateCmd)

	command.AddCommand(initCmd)
	command.AddCommand(joinCmd)

	command.AddCommand(ejectCmd)
	command.AddCommand(searchCmd)
	command.AddCommand(infoCmd)
	command.AddCommand(generateCmd)

	command.AddCommand(VersionCmd())

	command.PersistentFlags().BoolVarP(
		&flags.quiet, app.Trans("fleek.quietFlag"), "q", false, app.Trans("fleek.quietFlagDescription"))
	command.PersistentFlags().BoolVarP(
		&flags.verbose, app.Trans("fleek.verboseFlag"), "v", false, app.Trans("fleek.verboseFlagDescription"))

	debugMiddleware.AttachToFlag(command.PersistentFlags(), app.Trans("fleek.debugFlag"))
	traceMiddleware.AttachToFlag(command.PersistentFlags(), app.Trans("fleek.traceFlag"))

	return command
}

func mustConfig() error {

	if !cfgFound {
		return usererr.New("configuration files not found, run `fleek init`")
	}
	return nil
}
