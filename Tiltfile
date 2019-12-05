allow_k8s_contexts('kubernetes-admin@differ-cluster')

k8s_yaml(['local-dev/k8s/deployment.yaml','local-dev/k8s/rbac.yaml', 'local-dev/k8s/test-workload/appv1.yaml'])
docker_build('wiedmannfelix/differ', '.')

k8s_resource('differ-deployment', port_forwards='8080')