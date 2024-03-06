package main

import (
	"context"
	"fmt"
)

// TODO: rename GCR -> Artifact Registry (GAR?)

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

func (m *Gcp) GcpCli(ctx context.Context, gcpCredentials *File) (*Container, error) {
	ctr := dag.Container().
		From("gcr.io/google.com/cloudsdktool/google-cloud-cli:467.0.0")
	ctr, err := m.WithGcpSecret(ctx, ctr, gcpCredentials)
	if err != nil {
		return nil, err
	}
	return ctr, nil
}

// example usage: "dagger call list --account your@email.address --project gcp-project-id --gcp-credentials ~/.config/gcloud/credentials.db"
func (m *Gcp) List(ctx context.Context, account, project string, gcpCredentials *File) (string, error) {
	ctr, err := m.GcpCli(ctx, gcpCredentials)
	if err != nil {
		return "", err
	}
	return ctr.
		// WithEnvVariable("CLOUDSDK_CONFIG", "/tmp/.config/gcloud").
		WithEnvVariable("CLOUDSDK_CORE_PROJECT", project).
		// WithExec([]string{"mount"}).
		WithExec([]string{"gcloud", "--account", account, "compute", "instances", "list"}).
		Stdout(ctx)
}

// example usage: "dagger call gcr-get-login-password --region us-east-1 --gcp-credentials ~/.config/gcloud/credentials.db"
func (m *Gcp) GcrGetLoginPassword(ctx context.Context, gcpCredentials *File, region string) (string, error) {
	ctr, err := m.GcpCli(ctx, gcpCredentials)
	if err != nil {
		return "", err
	}
	// TODO: what is this?
	return ctr.
		WithExec([]string{"--region", region, "ecr", "get-login-password"}).
		Stdout(ctx)
}

// Push ubuntu:latest to GCR under given repo 'test' (repo must be created first)
// example usage: "dagger call gcr-push-example --region us-east-1 --gcp-credentials ~/.config/gcloud/credentials.db --gcp-account-id 12345 --repo test"
func (m *Gcp) GcrPushExample(ctx context.Context, gcpCredentials *File, region, gcpProject, repo, image string) (string, error) {
	ctr := dag.Container().From("ubuntu:latest")
	return m.GcrPush(ctx, gcpCredentials, region, gcpProject, repo, image, ctr)
}

func (m *Gcp) GcrPush(ctx context.Context, gcpCredentials *File, region, gcpProject, repo, image string, ctr *Container) (string, error) {
	// Get the GCR login password so we can authenticate with Publish WithRegistryAuth
	ctr, err := m.GcpCli(ctx, gcpCredentials)
	if err != nil {
		return "", err
	}
	/*

		PROJECT_ID="your-project-id"
		SERVICE_ACCOUNT_NAME="kube-registry-sa"
		DISPLAY_NAME="Kube Registry SA"

		gcloud iam service-accounts create $SERVICE_ACCOUNT_NAME \
			--display-name="$DISPLAY_NAME" \
			--project=$PROJECT_ID

		gcloud projects add-iam-policy-binding $PROJECT_ID \
			--member="serviceAccount:$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
			--role="roles/artifactregistry.reader"

		gcloud iam service-accounts keys create sa-key.json \
			--iam-account="$SERVICE_ACCOUNT_NAME@$PROJECT_ID.iam.gserviceaccount.com" \
			--project=$PROJECT_ID

	*/
	// TODO: what is this?
	regCred, err := ctr.
		WithExec([]string{"--region", region, "ecr", "get-login-password"}).
		Stdout(ctx)
	if err != nil {
		return "", err
	}
	// secret will be a service account json
	secret := dag.SetSecret("gcp-reg-cred", regCred)
	// region e.g. europe-west2
	gcrHost := fmt.Sprintf("%s-docker.pkg.dev", region)
	// e.g europe-west2-docker.pkg.dev/<project-id>/<artifact-repository>/<docker-image>
	gcrWithRepo := fmt.Sprintf("%s/%s/%s/%s", gcrHost, gcpProject, repo, image)

	return ctr.WithRegistryAuth(gcrHost, "_json_key_base64", secret).Publish(ctx, gcrWithRepo)
}
