# Dagger GCP module

## Push Image to Private Google Artifact Registry (GAR) Repo

This module lets you push container images from Dagger to GAR without having to manually configure docker with GAR credentials. It will allocate a service account and use the service account key to authenticate to google automatically behind the scenes, then clean up the service account.
