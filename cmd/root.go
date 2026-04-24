package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile   string
	vaultAddr string
	vaultToken string
)

var rootCmd = &cobra.Command{
	Use:   "vaultls",
	Short: "Browse and diff HashiCorp Vault secrets across environments",
	Long: `vaultls is a CLI tool for browsing, listing, and diffing
HashiCorp Vault secrets across multiple environments or paths.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $HOME/.vaultls.yaml)")
	rootCmd.PersistentFlags().StringVar(&vaultAddr, "addr", "", "Vault server address (overrides VAULT_ADDR)")
	rootCmd.PersistentFlags().StringVar(&vaultToken, "token", "", "Vault token (overrides VAULT_TOKEN)")

	viper.BindPFlag("vault.addr", rootCmd.PersistentFlags().Lookup("addr"))
	viper.BindPFlag("vault.token", rootCmd.PersistentFlags().Lookup("token"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		viper.AddConfigPath(home)
		viper.SetConfigName(".vaultls")
	}

	viper.SetEnvPrefix("VAULT")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
