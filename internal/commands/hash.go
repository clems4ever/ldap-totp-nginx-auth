package commands

import (
	"fmt"

	"github.com/simia-tech/crypt"
	"github.com/spf13/cobra"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
)

// NewHashPasswordCmd returns a new Hash Password Cmd.
func NewHashPasswordCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "hash-password [password]",
		Short: "Hash a password to be used in file-based users database. Default algorithm is argon2id.",
		Args:  cobra.MinimumNArgs(1),
		Run:   cmdHashPasswordRun,
	}

	cmd.Flags().BoolP("sha512", "z", false, fmt.Sprintf("use sha512 as the algorithm (changes iterations to %d, change with -i)", schema.DefaultPasswordSHA512Configuration.Iterations))
	cmd.Flags().IntP("iterations", "i", schema.DefaultPasswordConfiguration.Iterations, "set the number of hashing iterations")
	cmd.Flags().StringP("salt", "s", "", "set the salt string")
	cmd.Flags().IntP("memory", "m", schema.DefaultPasswordConfiguration.Memory, "[argon2id] set the amount of memory param (in MB)")
	cmd.Flags().IntP("parallelism", "p", schema.DefaultPasswordConfiguration.Parallelism, "[argon2id] set the parallelism param")
	cmd.Flags().IntP("key-length", "k", schema.DefaultPasswordConfiguration.KeyLength, "[argon2id] set the key length param")
	cmd.Flags().IntP("salt-length", "l", schema.DefaultPasswordConfiguration.SaltLength, "set the auto-generated salt length")

	return cmd
}

func cmdHashPasswordRun(cmd *cobra.Command, args []string) {
	sha512, _ := cmd.Flags().GetBool("sha512")
	iterations, _ := cmd.Flags().GetInt("iterations")
	salt, _ := cmd.Flags().GetString("salt")
	keyLength, _ := cmd.Flags().GetInt("key-length")
	saltLength, _ := cmd.Flags().GetInt("salt-length")
	memory, _ := cmd.Flags().GetInt("memory")
	parallelism, _ := cmd.Flags().GetInt("parallelism")

	var (
		hash      string
		algorithm authentication.CryptAlgo
	)

	if sha512 {
		if iterations == schema.DefaultPasswordConfiguration.Iterations {
			iterations = schema.DefaultPasswordSHA512Configuration.Iterations
		}

		algorithm = authentication.HashingAlgorithmSHA512
	} else {
		algorithm = authentication.HashingAlgorithmArgon2id
	}

	if salt != "" {
		salt = crypt.Base64Encoding.EncodeToString([]byte(salt))
	}

	hash, err := authentication.HashPassword(args[0], salt, algorithm, iterations, memory*1024, parallelism, keyLength, saltLength)
	if err != nil {
		logging.Logger().Fatalf("Error occurred during hashing: %v\n", err)
	}

	fmt.Printf("Password hash: %s\n", hash)
}
