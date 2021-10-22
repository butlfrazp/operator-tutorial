## Generating the Operator Template

In this lab, we will be utilizing `operator-sdk` to generate the resources we will need to create a Custom Resource (CR) and the corresponding Controller.
To do so, you will need to have the `operator-sdk` CLI installed from the previous lab.

**Note:** This operator will be built utilizing `golang`.
If you do not have `golang` installed, visit the first lab.

## Creating the Operator Infrastructure

### Creating the Project Directory

To generate the main scaffolding of the operator, the following commands will need to be run.

```bash
mkdir songs-operator
cd songs-operator
operator-sdk init \
    --domain <container_registry_url> \
    --repo github.com/<github_username>/songs-operator
```

The `init` command for operator-sdk creates the bare-bones project needed to create a new operator.
The domain flag expects a container registry url.
This url will be used to push the operator container on `make docker-build docker-push`.
The repo flag passed in will be the module name.

To generate a new CRD, run the following command.

```bash
operator-sdk create api \
    --group songs \
    --version v1alpha1 \
    --kind SongsConfig \
    --resource \
    --controller
```

The `resource` flag creates the resources needed to generate the CRD maifest file in the `api` directory.
The `controller` flag creates the base golang controller files in the `controller` directory.

### Understanding What's Generated

Here we are just going over some of the files you will be interfacing with.

- `songs-operator/api/v1alpha1/songsconfig_types.go`: Contains all the go structures of what the CRD should contain.
- `songs-operator/config/samples/songs_v1alpha1_songsconfig.yaml`: Sample implementation of the CRD
- `songs-operator/controllers/songsconfig_controller.go`: Where all the business logic of the operator will be. Ex: What happens on create or delete or update?