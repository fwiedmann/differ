# ! WIP !

<p align="center">
  <a href="https://example.com/">
    <img src="images/differ_logo.png" width=72 height=72>
  </a>

  <h3 align="center">differ</h3>

  <p align="center">
    Kubernetes controller differentiate images of running container in clusters namespaces against the source registry

    <br>
    <a href="https://github.com/fwiedmann/differ/releases/latest">Latest release </a>
    ·
        <a href="https://hub.docker.com/r/wiedmannfelix/differ">Docker Hub </a>
    ·
    <a href="https://github.com/fwiedmann/differ/issues/new?template=bug.md">Report bug</a>
    ·
    <a href="https://github.com/fwiedmann/differ/issues/new?template=feature.md&labels=feature">Request feature</a>
  </p>
</p>

  ![badge](https://action-badges.now.sh/fwiedmann/differ)

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
