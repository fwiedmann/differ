<p align="center">
  <a href="https://github.com/fwiedmann/differ">
    <img src="images/differ_logo.png" width=100 height=100>
  </a>

  <h3 align="center">Differ</h3>

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

# WIP

## ToDo's

### add tests

### add support for scanning github repos

### refectoring

- update all func and var names for clean code
- refactor functions to do one simple job like the Run Method
- use kubernetes watch method on resources to react on actions
- run registry in own routine
- reload config dynmaic like kubernetes


## ref



## RED metics

Rate

Errors
which images could not be scraped
images with now matched pattern

Duration
How long does a run needed