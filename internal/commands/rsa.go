package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/utils"
)

// NewRSACmd returns a new RSA Cmd.
func NewRSACmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "rsa",
		Short: "Commands related to rsa keypair generation",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(newRSAGenerateCmd())

	return cmd
}

func newRSAGenerateCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a RSA keypair",
		Args:  cobra.NoArgs,
		Run:   cmdRSAGenerateRun,
	}

	cmd.Flags().StringP("dir", "d", "", "Target directory where the keypair will be stored")
	cmd.Flags().IntP("key-size", "b", 2048, "Sets the key size in bits")

	return cmd
}

func cmdRSAGenerateRun(cmd *cobra.Command, _ []string) {
	bits, err := cmd.Flags().GetInt("key-size")
	if err != nil {
		fmt.Printf("Failed to parse key-size flag: %v\n", err)
		return
	}

	privateKey, publicKey := utils.GenerateRsaKeyPair(bits)

	rsaTargetDirectory, err := cmd.Flags().GetString("dir")
	if err != nil {
		fmt.Printf("Failed to parse dir flag: %v\n", err)
		return
	}

	keyPath := filepath.Join(rsaTargetDirectory, "key.pem")

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Failed to open %s for writing: %v\n", keyPath, err)
		return
	}

	defer func() {
		if err := keyOut.Close(); err != nil {
			fmt.Printf("Unable to close private key file: %v\n", err)
			os.Exit(1)
		}
	}()

	_, err = keyOut.WriteString(utils.ExportRsaPrivateKeyAsPemStr(privateKey))
	if err != nil {
		fmt.Printf("Failed to write private key: %v\n", err)
		return
	}

	fmt.Printf("RSA Private Key written to %s\n", keyPath)

	certPath := filepath.Join(rsaTargetDirectory, "key.pub")

	certOut, err := os.OpenFile(certPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Failed to open %s for writing: %v\n", keyPath, err)
		return
	}

	defer func() {
		if err := certOut.Close(); err != nil {
			fmt.Printf("Failed to close public key file: %v\n", err)
			os.Exit(1)
		}
	}()

	publicPem, err := utils.ExportRsaPublicKeyAsPemStr(publicKey)
	if err != nil {
		fmt.Printf("Failed to marshal public key: %v\n", err)
		return
	}

	_, err = certOut.WriteString(publicPem)
	if err != nil {
		fmt.Printf("Failed to write private key: %v\n", err)
		return
	}

	fmt.Printf("RSA Public Key written to %s\n", certPath)
}
