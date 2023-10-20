# Kai controller

Kai extends Kubneretes by providing higher level resources that make it easy to deploy and share your inference pipelines.

## Description
The Kai controller automates the management and integration of [kai-gateway](https://github.com/dreamstax/kai-gateway) which enables users to easily define and leverage ineference pipelines. 

Inference workloads typically involve more than just predictions. There's pre processing steps, potentially multiple predictions, and post processing steps. A variety of solutions exist to accomplish this but most are bloated, add unnecessary complexity, and are not portable. 

Kai aims to be a simple and minimal solution for defining and deploying inference pipelines whether it be for local use, research, or production environments.

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.

### Local installation [WIP]
1. Clone this repository
```sh 
git clone https://github.com/dreamstax/kai && cd kai
```

2. Download dependencies, create cluster, install CRDs, and deploy controller
```sh
make quickstart
```

This is an exhaustive build and install of all required dependencies and assumes nothing exists. This is mostly useful for initial repo pulls or starting from scratch. You can see a list of commands that are run within the makefile and perform the ones necessary during development.

3. After completion of the make command, ensure the controller is running in the cluster by checking the pods within the namespace.
```sh
kubectl get pods -n kai-system
```

4. Deploy example application
```sh
kubectl apply -f examples/http-echo/
```

5. Wait until resources become ready then port-forward gateway service
```sh
TODO: command
```

6. In a separate terminal curl the example application

```sh
curl http://localhost:8888/v1
```

### Running on the cluster
1. Install Custom Resources:

```sh
make install
```
	
2. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=quay.io/dreamstax/kai-controller:latest
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
Undeploy the controller from the cluster:

```sh
make undeploy
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

### How it works [WIP]
This project uses kubebuilder to scaffold the creation of kubernetes resources and controllers. Kai defines the following custom resources:


#### Example pipeline.yaml
TODO: example pipeline yaml

### Project Goals
This project draws inspiration from both KNative and KServe. It hopes to provide a modern and simple solution to the same problems those projects hope to solve.

Goals:
- Minimal dependencies
- UX
- Portable

### Project status
This project is currently in early alpha. Deployment to production is not recommended. As a project in early alpha you can expect frequent breaking and non-breaking changes.

We hope to gather community feedback around all aspects of the project both current and future.

## Contributing
If you would like to provide any feedback or have suggestions for new features please open an issue.

### Who is dreamstax...
Dreamstax is a group of individuals dedicated to opensource and making AI accessible
For more info about the team please visit [dreamstax.io](https://dreamstax.io)

