package commands

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/open-policy-agent/conftest/downloader"
	"github.com/sigstore/cosign/pkg/oci/remote"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	orascontext "oras.land/oras-go/pkg/context"
)

const pullDesc = `
This command downloads individual policies from a remote location.

Several locations are supported by the pull command. Under the hood
conftest leverages go-getter (https://github.com/hashicorp/go-getter).
The following protocols are supported for downloading policies:

	- OCI Registries
	- Local Files
	- Git
	- HTTP/HTTPS
	- Mercurial
	- Amazon S3
	- Google Cloud Storage

The location of the policies is specified by passing an URL, e.g.:

	$ conftest pull http://<my-policy-url>

Based on the protocol a different mechanism will be used to download the policy.
The pull command will also try to infer the protocol based on the URL if the 
URL does not contain a protocol. For example, the OCI mechanism will be used if
an azure registry URL is passed, e.g.

	$ conftest pull instrumenta.azurecr.io/my-registry

The policy location defaults to the policy directory in the local folder.
The location can be overridden with the '--policy' flag, e.g.:

	$ conftest pull --policy <my-directory> <oci-url>
`

// NewPullCommand creates a new pull command to allow users
// to download individual policies.
func NewPullCommand(ctx context.Context) *cobra.Command {
	cmd := cobra.Command{
		Use:   "pull <repository>",
		Short: "Download individual policies",
		Long:  pullDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if err := viper.BindPFlag("policy", cmd.Flags().Lookup("policy")); err != nil {
				return fmt.Errorf("bind flag: %w", err)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				cmd.Usage() //nolint
				return fmt.Errorf("missing required arguments")
			}

			ctx = orascontext.Background()

			if isVerify, err := cmd.Flags().GetString("verify"); err == nil && isVerify == "cosign" {
				for _, url := range args {
					t, err := downloader.Detect(url, "")
					if err != nil {
						return fmt.Errorf("detect type: %w", err)
					}
					if strings.HasPrefix(t, "oci://") {
						keyRef, err := cmd.Flags().GetString("cosign-key")
						if err != nil {
							return err
						}
						if err := verifyCosign(ctx, url, keyRef); err != nil {
							return err
						}
					}
				}
			}

			policyDir := filepath.Join(".", viper.GetString("policy"))

			if err := downloader.Download(ctx, policyDir, args); err != nil {
				return fmt.Errorf("download policies: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringP("policy", "p", "policy", "Path to download the policies to")
	cmd.Flags().String("verify", "none", "Verify the image with none|cosign. Default none")
	cmd.Flags().String("cosign-key", "", "path to the public key file, KMS, URI or Kubernetes Secret")

	return &cmd
}

func verifyCosign(ctx context.Context, rawRef string, keyRef string) error {
	ref, err := name.ParseReference(rawRef)
	if err != nil {
		return err
	}

	digest, err := remote.ResolveDigest(ref)
	if err != nil {
		return err
	}

	if err != nil {
		return fmt.Errorf("unable to resolve digest for an image %s: %v\n", digest.String(), err)
	}

	log.Printf("verifying image: %s\n", digest.String())

	cosignExecutable, err := exec.LookPath("cosign")
	if err != nil {
		return fmt.Errorf("cosign executable not found in path $PATH")
	}

	cosignCmd := exec.CommandContext(ctx, cosignExecutable, []string{"verify"}...)
	cosignCmd.Env = os.Environ()

	if keyRef != "" {
		cosignCmd.Args = append(cosignCmd.Args, "--key", keyRef)
	} else {
		cosignCmd.Env = append(cosignCmd.Env, "COSIGN_EXPERIMENTAL=true")
	}

	cosignCmd.Args = append(cosignCmd.Args, digest.String())

	log.Printf("running %s %v\n", cosignExecutable, cosignCmd.Args)

	stdout, _ := cosignCmd.StdoutPipe()
	stderr, _ := cosignCmd.StderrPipe()
	if err := cosignCmd.Start(); err != nil {
		return err
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		log.Println("cosign: " + scanner.Text())
	}

	errScanner := bufio.NewScanner(stderr)
	for errScanner.Scan() {
		log.Println("cosign: " + errScanner.Text())
	}

	return cosignCmd.Wait()
}
