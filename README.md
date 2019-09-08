# ! WIP !

# differ

Kubernetes controller differentiate images of running container in clusters namespaces against the source registry

## Concept

### Features

-   Configuration via ConfigMap
    -   Configure for 1 kubernetes namespaces
    -   Enable metrics y/n
    -   Exclude images
    -   regex pattern for exclude image versions
    -   Kubernetes container registry secret name for private registry's
    -   Git repository links and kubernetes secret name which provides the API token
        -   folders/files to ignore in repo
-   Watch running pods in configured namespaces
    -   Get all containers in pod
        -   Get container image name
        -   Get Container image
-   Get latest image version of containers from source container image registry
-   Expose metrics for each differences between container image versions
-   Update configured git repository's and their kubernetes manifests to new image version
