# Dagger GCP module

Known to work with v0.9.8.

## Push Image to Private Google Artifact Registry (GAR) Repo

This module lets you push container images from Dagger to GAR without having to manually configure docker with GAR credentials. It will fetch them automatically behind the scenes.

From CLI, to push `ubuntu:latest` to a given GAR repo, by way of example:

```
dagger call gar-push-example \
    --region us-east-1 --gcp-credentials ~/.config/gcloud/credentials.db \
    --gcp-account-id 12345 --repo test
```
