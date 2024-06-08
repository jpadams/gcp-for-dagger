// Push a container image into Google Artifact Registry
//
// This module lets you push a container into Google Artifact Registry, automating the tedious manual steps of setting up a service account for the docker credential

package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

type Gcp struct{}

func (m *Gcp) GetSecret(ctx context.Context, gcpCredentials *File) (string, error) {
	ctr, err := m.WithGcpSecret(ctx, dag.Container().From("ubuntu:latest"), gcpCredentials)
	if err != nil {
		return "", err
	}
	return ctr.
		WithExec([]string{"bash", "-c", "cat /root/.config/gcloud/credentials.db |base64"}).
		Stdout(ctx)
}

func (m *Gcp) WithGcpSecret(ctx context.Context, ctr *Container, gcpCredentials *File) (*Container, error) {
	// gcloud wants to open file as writable, so we can't use WithMountedSecret here sadly
	return ctr.WithFile("/root/.config/gcloud/credentials.db", gcpCredentials), nil

	// credsFile, err := gcpCredentials.Contents(ctx)
	// if err != nil {
	// 	return nil, err
	// }
	// secret := dag.SetSecret("gcp-credential", credsFile)
	// return ctr.WithMountedSecret("/tmp/.config/gcloud/credentials.db", secret), nil
}

func (m *Gcp) GcloudCli(ctx context.Context, project string, gcpCredentials *File) (*Container, error) {
	ctr := dag.Container().
		From("gcr.io/google.com/cloudsdktool/google-cloud-cli:467.0.0").
		WithEnvVariable("BUST_CACHE", time.Now().String()).
		WithEnvVariable("CLOUDSDK_CORE_PROJECT", project)
	ctr, err := m.WithGcpSecret(ctx, ctr, gcpCredentials)
	if err != nil {
		return nil, err
	}
	return ctr, nil
}

func (m *Gcp) List(ctx context.Context, account, project string, gcpCredentials *File) (string, error) {
	ctr, err := m.GcloudCli(ctx, project, gcpCredentials)
	if err != nil {
		return "", err
	}
	return ctr.
		// WithEnvVariable("CLOUDSDK_CONFIG", "/tmp/.config/gcloud").
		// WithExec([]string{"mount"}).
		WithExec([]string{"gcloud", "--account", account, "compute", "instances", "list"}).
		Stdout(ctx)
}

func (m *Gcp) GarEnsureServiceAccountKey(ctx context.Context, account, region, project string, gcpCredentials *File) (string, error) {
	ctr, err := m.GcloudCli(ctx, project, gcpCredentials)
	if err != nil {
		return "", err
	}
	// We create a temporary service account key, because we can't persist it to
	// the host, and don't want to burden the user with it via download etc.
	// TODO: maybe use cache volumes to persist a service account between
	// invocations? Maybe not super secure though.
	// No, short lived keys are a feature :-)

	saName := "dagger-image-push"
	saDisplayName := "Push artifact registry images from Dagger"
	saFullName := fmt.Sprintf("serviceAccount:%s@%s.iam.gserviceaccount.com", saName, project)
	saShortName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, project)

	args := []string{"gcloud", "--account", account,
		"iam", "service-accounts", "get-iam-policy", saShortName,
	}
	_, err = ctr.
		WithExec(args).
		Stdout(ctx)
	if err != nil {
		// XXX Sad that we can't differentiate between a real permission denied
		// and a "policy does not exist". Maybe list with a filter (if that's a
		// thing) would be better?
		if strings.Contains(fmt.Sprintf("%s", err), "PERMISSION_DENIED") {
			// policy might not exist, so create it
			_, err := ctr.
				WithExec([]string{"gcloud", "--account", account,
					"iam", "service-accounts", "create", saName,
					fmt.Sprintf(`--display-name=%s`, saDisplayName),
				}).
				Stdout(ctx)

			if err != nil {
				// create failed
				return "", err
			}
		} else {
			// some other error getting policy, propagate it
			return "", err
		}
	}
	// at this point, service account will exist.
	// we set the role on the project every time because setting it twice seems
	// to be a noop and this allows us to recover from the case where this
	// processes was previously interrupted midway through.

	_, err = ctr.
		WithExec([]string{"gcloud", "--account", account,
			"projects", "add-iam-policy-binding", project,
			"--member", saFullName,
			`--role=roles/artifactregistry.admin`,
		}).
		Stdout(ctx)
	if err != nil {
		log.Printf("error adding iam policy binding: %s", err)
		return "", err
	}

	// now create a single-use sa key that we can use for this push.
	// The calling function is responsible for cleaning us up.
	ctr = ctr.
		WithExec([]string{
			"gcloud", "--account", account, "iam", "service-accounts", "keys",
			"create", "/tmp/sa-key.json", "--iam-account", saShortName,
		})
	_, err = ctr.Stdout(ctx)
	if err != nil {
		log.Printf("error creating key for sa: %s", err)
		return "", err
	}
	return ctr.File("/tmp/sa-key.json").Contents(ctx)
}

type ServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

// Push ubuntu:latest to GAR under given repo 'test' (repo must be created first)
func (m *Gcp) GarPushExample(ctx context.Context, account, region, project, repo, image string, gcpCredentials *File) (string, error) {
	ctr := dag.Container().From("ubuntu:latest")
	return m.GarPush(ctx, ctr, account, region, project, repo, image, gcpCredentials)
}

func (m *Gcp) CleanupServiceAccountKey(ctx context.Context, account, region, project string, gcpCredentials *File, keyId string) error {
	saName := "dagger-image-push"
	saShortName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, project)
	ctr, err := m.GcloudCli(ctx, project, gcpCredentials)
	if err != nil {
		log.Printf("error getting gcloud cli in up key id %s: %s", keyId, err)
		return err
	}
	_, err = ctr.
		WithExec([]string{
			"gcloud", "--account", account,
			"iam", "service-accounts", "keys", "delete",
			keyId, "--iam-account", saShortName,
		}).
		Stdout(ctx)
	if err != nil {
		log.Printf("error cleaning up key id %s: %s", keyId, err)
		return err
	}
	return nil
}

func (m *Gcp) GarPush(ctx context.Context, pushCtr *Container, account, region, project, repo, image string, gcpCredentials *File) (string, error) {
	// Get the GAR login password so we can authenticate with Publish WithRegistryAuth
	saStr, err := m.GarEnsureServiceAccountKey(ctx, account, region, project, gcpCredentials)
	if err != nil {
		log.Printf("error ensuring key for sa: %s", err)
		return "", err
	}
	sa := &ServiceAccount{}
	err = json.Unmarshal([]byte(saStr), sa)
	if err != nil {
		log.Printf("error unmarshalling key for sa: %s", err)
		return "", err
	}

	defer m.CleanupServiceAccountKey(ctx, account, region, project, gcpCredentials, sa.PrivateKeyID)
	bs := base64.StdEncoding.EncodeToString([]byte(saStr))

	// secret will be a service account json
	secret := dag.SetSecret("gcp-reg-cred", bs)
	// region e.g. europe-west2
	garHost := fmt.Sprintf("%s-docker.pkg.dev", region)
	// e.g europe-west2-docker.pkg.dev/<project-id>/<artifact-repository>/<docker-image>
	garWithRepo := fmt.Sprintf("%s/%s/%s/%s", garHost, project, repo, image)

	return pushCtr.WithRegistryAuth(garHost, "_json_key_base64", secret).Publish(ctx, garWithRepo)
}
