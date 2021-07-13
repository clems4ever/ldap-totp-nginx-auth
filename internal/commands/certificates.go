package commands

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// NewCertificatesCmd returns a new Certificates Cmd.
func NewCertificatesCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "certificates",
		Short: "Commands related to certificate generation",
		Args:  cobra.NoArgs,
	}

	cmd.PersistentFlags().StringSlice("host", []string{}, "Comma-separated hostnames and IPs to generate a certificate for")

	err := cmd.MarkPersistentFlagRequired("host")
	if err != nil {
		log.Fatal(err)
	}

	cmd.AddCommand(newCertificatesGenerateCmd())

	return cmd
}

func newCertificatesGenerateCmd() (cmd *cobra.Command) {
	cmd = &cobra.Command{
		Use:   "generate",
		Short: "Generate a self-signed certificate",
		Args:  cobra.NoArgs,
		Run:   cmdCertificatesGenerateRun,
	}

	cmd.Flags().String("start-date", "", "Creation date formatted as Jan 1 15:04:05 2011")
	cmd.Flags().Duration("duration", 365*24*time.Hour, "Duration that certificate is valid for")
	cmd.Flags().Bool("ca", false, "Whether this cert should be its own Certificate Authority")
	cmd.Flags().Int("rsa-bits", 2048, "Size of RSA key to generate. Ignored if --ecdsa-curve is set")
	cmd.Flags().String("ecdsa-curve", "", "ECDSA curve to use to generate a key. Valid values are P224, P256 (recommended), P384, P521")
	cmd.Flags().Bool("ed25519", false, "Generate an Ed25519 key")
	cmd.Flags().String("dir", "", "Target directory where the certificate and keys will be stored")

	return cmd
}

func cmdCertificatesGenerateRun(cmd *cobra.Command, _ []string) {
	// implementation retrieved from https://golang.org/src/crypto/tls/generate_cert.go
	ecdsaCurve, err := cmd.Flags().GetString("ecdsa-curve")
	if err != nil {
		fmt.Printf("Failed to parse ecdsa-curve flag: %v\n", err)
		os.Exit(1)
	}

	ed25519Key, err := cmd.Flags().GetBool("ed25519")
	if err != nil {
		fmt.Printf("Failed to parse ed25519 flag: %v\n", err)
		os.Exit(1)
	}

	rsaBits, err := cmd.Flags().GetInt("rsa-bits")
	if err != nil {
		fmt.Printf("Failed to parse rsa-bits flag: %v\n", err)
		os.Exit(1)
	}

	hosts, err := cmd.Flags().GetStringSlice("host")
	if err != nil {
		fmt.Printf("Failed to parse host flag: %v\n", err)
		os.Exit(1)
	}

	validFrom, err := cmd.Flags().GetString("start-date")
	if err != nil {
		fmt.Printf("Failed to parse start-date flag: %v\n", err)
		os.Exit(1)
	}

	validFor, err := cmd.Flags().GetDuration("duration")
	if err != nil {
		fmt.Printf("Failed to parse duration flag: %v\n", err)
		os.Exit(1)
	}

	isCA, err := cmd.Flags().GetBool("ca")
	if err != nil {
		fmt.Printf("Failed to parse ca flag: %v\n", err)
		os.Exit(1)
	}

	certificateTargetDirectory, err := cmd.Flags().GetString("dir")
	if err != nil {
		fmt.Printf("Failed to parse dir flag: %v\n", err)
		os.Exit(1)
	}

	cmdCertificatesGenerateRunExtended(hosts, ecdsaCurve, validFrom, certificateTargetDirectory, ed25519Key, isCA, rsaBits, validFor)
}

func cmdCertificatesGenerateRunExtended(hosts []string, ecdsaCurve, validFrom, certificateTargetDirectory string, ed25519Key, isCA bool, rsaBits int, validFor time.Duration) {
	priv, err := getPrivateKey(ecdsaCurve, ed25519Key, rsaBits)

	if err != nil {
		fmt.Printf("Failed to generate private key: %v\n", err)
		os.Exit(1)
	}

	var notBefore time.Time

	switch len(validFrom) {
	case 0:
		notBefore = time.Now()
	default:
		notBefore, err = time.Parse("Jan 2 15:04:05 2006", validFrom)
		if err != nil {
			fmt.Printf("Failed to parse start date: %v\n", err)
			os.Exit(1)
		}
	}

	notAfter := notBefore.Add(validFor)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)

	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		fmt.Printf("Failed to generate serial number: %v\n", err)
		os.Exit(1)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Acme Co"},
		},
		NotBefore: notBefore,
		NotAfter:  notAfter,

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	if isCA {
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
	}

	certPath := filepath.Join(certificateTargetDirectory, "cert.pem")

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey(priv), priv)
	if err != nil {
		fmt.Printf("Failed to create certificate: %v\n", err)
		os.Exit(1)
	}

	writePEM(derBytes, "CERTIFICATE", certPath)

	fmt.Printf("Certificate Public Key written to %s\n", certPath)

	keyPath := filepath.Join(certificateTargetDirectory, "key.pem")

	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		fmt.Printf("Failed to marshal private key: %v\n", err)
		os.Exit(1)
	}

	writePEM(privBytes, "PRIVATE KEY", keyPath)

	fmt.Printf("Certificate Private Key written to %s\n", keyPath)
}

func getPrivateKey(ecdsaCurve string, ed25519Key bool, rsaBits int) (priv interface{}, err error) {
	switch ecdsaCurve {
	case "":
		if ed25519Key {
			_, priv, err = ed25519.GenerateKey(rand.Reader)
		} else {
			priv, err = rsa.GenerateKey(rand.Reader, rsaBits)
		}
	case "P224":
		priv, err = ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
	case "P256":
		priv, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	case "P384":
		priv, err = ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	case "P521":
		priv, err = ecdsa.GenerateKey(elliptic.P521(), rand.Reader)
	default:
		err = fmt.Errorf("unrecognized elliptic curve: %q", ecdsaCurve)
	}

	return priv, err
}

func writePEM(bytes []byte, blockType, path string) {
	keyOut, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		fmt.Printf("Failed to open %s for writing: %v\n", path, err)
		os.Exit(1)
	}

	if err := pem.Encode(keyOut, &pem.Block{Type: blockType, Bytes: bytes}); err != nil {
		fmt.Printf("Failed to write data to %s: %v\n", path, err)
		os.Exit(1)
	}

	if err := keyOut.Close(); err != nil {
		fmt.Printf("Error closing %s: %v\n", path, err)
		os.Exit(1)
	}
}

func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	case ed25519.PrivateKey:
		return k.Public().(ed25519.PublicKey)
	default:
		return nil
	}
}
