package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"carto/config"
	"carto/doctor"
	"carto/exitcode"
	"carto/schema"
	"carto/skill"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor [path]",
	Short: "Diagnose whether carto output can be trusted for a repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}
		report := doctor.Run(root)
		if jsonOut {
			out, _ := json.MarshalIndent(report, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Print(report.String())
		}
		if report.Status == "error" {
			return fmt.Errorf("%w: doctor found blocking problems", exitcode.ErrIO)
		}
		return nil
	},
}

var schemaCmd = &cobra.Command{
	Use:   "schema [command]",
	Short: "Print the JSON output contract for a command",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			names := schema.List()
			sort.Strings(names)
			if jsonOut {
				out, _ := json.MarshalIndent(map[string]interface{}{"schema_version": 1, "commands": names}, "", "  ")
				fmt.Println(string(out))
			} else {
				fmt.Println("Commands with a published JSON contract:")
				for _, name := range names {
					fmt.Printf("  %s\n", name)
				}
			}
			return nil
		}
		definition, ok := schema.Get(args[0])
		if !ok {
			return fmt.Errorf("no schema for %q; run 'carto schema' to list commands", args[0])
		}
		out, _ := json.MarshalIndent(definition, "", "  ")
		fmt.Println(string(out))
		return nil
	},
}

var initCmd = &cobra.Command{
	Use:   "init [path]",
	Short: "Initialize carto config and CLI guidance in a repository",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := "."
		if len(args) > 0 {
			root = args[0]
		}
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return err
		}
		force, _ := cmd.Flags().GetBool("force")
		skipConfig, _ := cmd.Flags().GetBool("skip-config")
		skipAgents, _ := cmd.Flags().GetBool("skip-agents")
		created := []string{}
		if !skipConfig {
			path := config.Path(absRoot)
			if _, err := os.Stat(path); os.IsNotExist(err) || force {
				if err := config.Save(absRoot, config.Default()); err != nil {
					return err
				}
				created = append(created, path)
			}
		}
		if !skipAgents {
			files, err := skill.Install(absRoot, "both")
			if err != nil {
				return err
			}
			created = append(created, files...)
		}
		if jsonOut {
			out, _ := json.MarshalIndent(map[string]interface{}{"written": created}, "", "  ")
			fmt.Println(string(out))
		} else {
			for _, f := range created {
				fmt.Printf("Wrote %s\n", f)
			}
		}
		return nil
	},
}

func init() {
	initCmd.Flags().Bool("force", false, "Overwrite an existing .carto/config.json")
	initCmd.Flags().Bool("skip-config", false, "Do not create .carto/config.json")
	initCmd.Flags().Bool("skip-agents", false, "Do not install the skill or AGENTS.md guide")
}

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "Print or install the carto CLI skill",
	RunE: func(cmd *cobra.Command, args []string) error {
		install, _ := cmd.Flags().GetBool("install")
		check, _ := cmd.Flags().GetBool("check")
		target, _ := cmd.Flags().GetString("target")
		root, _ := cmd.Flags().GetString("root")
		if target != "claude" && target != "agents" && target != "both" {
			return fmt.Errorf("target must be claude, agents, or both")
		}
		absRoot, err := filepath.Abs(root)
		if err != nil {
			return err
		}
		if install && check {
			return fmt.Errorf("--install and --check cannot be combined")
		}
		if install {
			files, err := skill.Install(absRoot, target)
			if err != nil {
				return err
			}
			if jsonOut {
				out, _ := json.MarshalIndent(map[string]interface{}{"installed": files}, "", "  ")
				fmt.Println(string(out))
			} else {
				for _, f := range files {
					fmt.Printf("Installed %s\n", f)
				}
			}
			return nil
		}
		if check {
			problems := skill.Check(absRoot, target)
			if jsonOut {
				out, _ := json.MarshalIndent(map[string]interface{}{"current": len(problems) == 0, "problems": problems}, "", "  ")
				fmt.Println(string(out))
			} else if len(problems) == 0 {
				fmt.Println("CLI skill is current.")
			} else {
				for _, p := range problems {
					fmt.Printf("Skill drift: %s\n", p)
				}
			}
			if len(problems) > 0 {
				return fmt.Errorf("%w: CLI skill is out of date; run 'carto skill --install'", exitcode.ErrFindings)
			}
			return nil
		}
		if jsonOut {
			out, _ := json.MarshalIndent(map[string]interface{}{"name": "carto", "skill": skill.Text}, "", "  ")
			fmt.Println(string(out))
		} else {
			fmt.Print(skill.Text)
		}
		return nil
	},
}

func init() {
	skillCmd.Flags().Bool("install", false, "Install the skill and AGENTS.md guide")
	skillCmd.Flags().Bool("check", false, "Check installed copies for drift")
	skillCmd.Flags().String("target", "both", "Install/check target: claude, agents, or both")
	skillCmd.Flags().String("root", ".", "Project root for install/check")
}
