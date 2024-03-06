package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// TODO: rename GAR -> Artifact Registry (GAR?)

type Gcp struct{}

// example usage: "dagger call get-secret --gcp-credentials ~/.config/gcloud/credentials.db"
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

// example usage: "dagger call list --account your@email.address --project gcp-project-id --gcp-credentials ~/.config/gcloud/credentials.db"
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

// example usage: "dagger call gar-get-service-account --region us-east-1 --project gcp-project-id --gcp-credentials ~/.config/gcloud/credentials.db"
func (m *Gcp) GarGetServiceAccount(ctx context.Context, account, region, project string, gcpCredentials *File) (string, error) {
	ctr, err := m.GcloudCli(ctx, project, gcpCredentials)
	if err != nil {
		return "", err
	}
	// We create a temporary service account key, because we can't persist it to
	// the host, and don't want to burden the user with it via download etc.
	// TODO: maybe use cache volumes to persist a service account between
	// invocations? Maybe not super secure though.

	/*

		PROJECT_ID="your-project-id"
		SERVICE_ACCOUNT_NAME="kube-registry-sa"
		DISPLAY_NAME="Kube Registry SA"

		gcloud iam service-accounts create $SERVICE_ACCOUNT_NAME \
			--display-name="$DISPLAY_NAME"

		gcloud projects add-iam-policy-binding $PROJECT_ID \
			--member="serviceAccount:$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
			--role="roles/artifactregistry.writer"

		gcloud iam service-accounts keys create sa-key.json \
			--iam-account="$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com" \

	*/
	saName := "dagger-image-push"
	saDisplayName := "Push artifact registry images from Dagger"
	// saFullName := fmt.Sprintf("serviceAccount:%s@%s.iam.gserviceaccount.com", saName, project)
	saShortName := fmt.Sprintf("%s@%s.iam.gserviceaccount.com", saName, project)

	args := []string{"gcloud", "--account", account,
		"iam", "service-accounts", "get-iam-policy", saShortName,
	}
	_, err = ctr.
		WithExec(args).
		Stdout(ctx)
	if err != nil {
		if strings.Contains(fmt.Sprintf("%s", err), "PERMISSION_DENIED") {
			// policy might not exist, so create it
			_, err := ctr.
				WithExec([]string{"gcloud", "--account", account,
					"iam", "service-accounts", "create", saName,
					fmt.Sprintf(`--display-name=%s`, saDisplayName),
				}).
				Stdout(ctx)

			if err != nil {
				return "", err
			}
		} else {
			// some other error, propagate it
			return "", err
		}
	}
	// at this point, service account will exist

	stdout, err := ctr.
		WithExec([]string{"gcloud", "--account", account,
			"iam", "service-accounts", "add-iam-policy-binding", project,
			fmt.Sprintf(`--member=%s`, saShortName),
			`--role=roles/artifactregistry.writer`,
		}).
		Stdout(ctx)

	fmt.Printf("stdout 2: %s\n", stdout)

	return "TODO", nil
}

// Push ubuntu:latest to GAR under given repo 'test' (repo must be created first)
// example usage: "dagger call gar-push-example --region us-east-1 --gcp-credentials ~/.config/gcloud/credentials.db --gcp-account-id 12345 --repo test"
func (m *Gcp) GarPushExample(ctx context.Context, gcpCredentials *File, account, region, project, repo, image string) (string, error) {
	ctr := dag.Container().From("ubuntu:latest")
	return m.GarPush(ctx, gcpCredentials, account, region, project, repo, image, ctr)
}

func (m *Gcp) GarPush(ctx context.Context, gcpCredentials *File, account, region, project, repo, image string, pushCtr *Container) (string, error) {
	// Get the GAR login password so we can authenticate with Publish WithRegistryAuth
	sa, err := m.GarGetServiceAccount(ctx, account, region, project, gcpCredentials)
	if err != nil {
		return "", err
	}
	// secret will be a service account json
	secret := dag.SetSecret("gcp-reg-cred", sa)
	// region e.g. europe-west2
	garHost := fmt.Sprintf("%s-docker.pkg.dev", region)
	// e.g europe-west2-docker.pkg.dev/<project-id>/<artifact-repository>/<docker-image>
	garWithRepo := fmt.Sprintf("%s/%s/%s/%s", garHost, project, repo, image)

	return pushCtr.WithRegistryAuth(garHost, "_json_key_base64", secret).Publish(ctx, garWithRepo)
}
