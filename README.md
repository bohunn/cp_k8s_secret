# Kubernetes Secret Watcher
This is a Go program that watches Kubernetes secrets and performs certain actions based on the type of the event.

## Prerequisites
This program requires the following tools to be installed:

- Go
- Kubernetes cluster

## Usage
To use this program, you must either build locally the Docker image from Dockerfile.

Or use the image present in the Docker Hub under ```bohunn/cp-k8s-secret```

Next, you need to set up a configuration file. The configuration file should be in YAML format and contain the following fields:

secret_name: Comma-separated list of the names of the secrets to watch
namespace: Namespace to watch for secrets
deletion_policy: Policy to follow when a secret is deleted. The value can be ORPHAN or DELETE
An example configuration file might look like this:

```vbnet
Copy code
secret_name: secret1,secret2
namespace: default
deletion_policy: DELETE
```

Once created mount it to your Deployment under ```/config.cfg```

All example .yaml files for deployment on Kubernetes can be found in repo ```deploy/``` order.

## Actions
### Create Secret
If a new secret is added to the namespace, the program will create a new secret with the same name in the target namespace. If a secret with the same name already exists in the target namespace, the program will update the existing secret.

### Update Secret
If a secret is modified, the program will update the corresponding secret in the target namespace.

### Delete Secret
If a secret is deleted, the program will delete the corresponding secret in the target namespace. The deletion policy is determined by the deletion_policy field in the configuration file. If the policy is set to DELETE, the secret will be deleted. If the policy is set to ORPHAN, the program will do nothing.

## License

Kubernetes Secret Watcher is released under the MIT License.

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
