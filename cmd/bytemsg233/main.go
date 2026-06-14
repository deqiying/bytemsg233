package main

import (
	"fmt"
	"os"

	"github.com/neko233-com/bytemsg233/pkg/compiler"
	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "bytemsg233",
		Short: "bytemsg233 - A modern serialization framework",
		Long:  "bytemsg233 - A modern serialization framework that replaces Protocol Buffers",
	}

	var compileCmd = &cobra.Command{
		Use:   "compile [file]",
		Short: "Compile .bmsg or .bmsg.yaml to target languages",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			languages, _ := cmd.Flags().GetStringSlice("lang")
			outputDir, _ := cmd.Flags().GetString("output")
			locale, _ := cmd.Flags().GetString("locale")

			comp := compiler.New()
			return comp.Compile(&compiler.CompileOptions{
				InputFile: args[0],
				OutputDir: outputDir,
				Languages: languages,
				Locale:    locale,
			})
		},
	}

	compileCmd.Flags().StringSliceP("lang", "l", []string{"go"}, "Target languages (go, csharp, java, typescript, python)")
	compileCmd.Flags().StringP("output", "o", ".", "Output directory")
	compileCmd.Flags().String("locale", "en", "Locale for comments (en, zh)")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("bytemsg233 %s (commit: %s, built: %s)\n", version, commit, date)
		},
	}

	var initCmd = &cobra.Command{
		Use:   "init [name]",
		Short: "Initialize a new .bmsg file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			template := fmt.Sprintf(`schema: bymsg/v1
package: %s

enum Status {
    ACTIVE = 0
    INACTIVE = 1
}

message Example {
    uint32 id = 1 // "ID" | "ID"
    string name = 2 // "名称" | "Name"
}
`, name)

			filename := fmt.Sprintf("%s.bmsg", name)
			return os.WriteFile(filename, []byte(template), 0644)
		},
	}

	rootCmd.AddCommand(compileCmd, versionCmd, initCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
