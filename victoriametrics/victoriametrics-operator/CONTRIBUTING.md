
## Required programs

for developing you need: 
- golang 1.13+
- operator-sdk v1.0.0
- docker
- minikube or kind for e2e tests
- golangci-lint



## local build and run

Use `make build` - it will generate new crds and build binary


for running locally you need minikube and run two commands:
```bash
make install
make run
```
or you can run it from IDE with ```main.go```

## publish changes

before creating merge request, ensure that tests passed locally:
```bash
make build # it will update crds
make lint # linting project
make test #unit tests
make e2e-local #e2e tests with minikube
```

## adding new api

For adding new kind - KIND_NAME, you have to execute command:

```bash
operator-sdk create api --group operator --version v1beta1 --kind KIND_NAME
```

This will scaffold api and controller. Then you have to edit code at `api` and `controllers` folder.

## create olm package

Choose version and generate or update corresponding csv file
```bash
CSV_VERSION=0.0.2 make olm
```

it will generate files at directory: `deploy/olm-catalog/victoria-metrics-operator/0.0.3/`

update file: `deploy/olm-catalog/victoria-metrics-operator/0.0.3/victoria-metrics-operator.v0.0.3.clusterserviceversion.yaml`

with fields from file:`deploy/olm-catalog/templates/csv-additional-fields-template.yaml`

replace operator version - specify release instead of latest

commit changes

publish olm package to quay.io with

```bash
make olm-publish
```

### integration with operator-hub

 Clone repo locally: git clone https://github.com/operator-framework/community-operators.git
 
 copy content to operator-hub repo and run tests
 you can specify version (OP_VER) and channel OP_CHANNEL
 ```bash
cp -R deploy/olm-catalog/victoria-metrics-operator/ PATH_TO_OPERATOR_REPO/upstream-community-operators/
cd PATH_TO_OPERATOR_REPO
#run tests
make operator.verify OP_PATH=upstream-community-operators/victoria-metrics-operator VERBOSE=1
make operator.test OP_PATH=upstream-community-operators/victoria-metrics-operator/ VERBOSE=1

```

 Now you can submit merge request with changes to operator-hub repo


troubleshooting: [url](https://github.com/operator-framework/community-operators/blob/master/docs/using-scripts.md#troubleshooting)