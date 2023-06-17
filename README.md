# Kai controller

Kai extends Kubernetes by providing higher level resources that support deploying and managing applications. Kai makes it easy to deploy applications to Kubernetes while utilizing best practices and strategies.

## Description
The Kai project provides the following features:
- Easy application deployment
- Application networking and routing
- Automatic scaling
- Immutable versions for varying rollout strategies, and rollbacks
- Easy installation: requires kubernetes >= `1.24` && [Kubernetes Gateway](https://gateway-api.sigs.k8s.io/) controller [implementation](https://gateway-api.sigs.k8s.io/implementations/)

Future goals:
- Provide easy deployment of ML Models while relying on core kai resources
- Provide supporting resources for managing and scaling ML Models

## Getting Started
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** The controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Local installation
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
kubectl -n projectcontour port-forward service/envoy-kai-gateway 8888:8080
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
make deploy IMG=<some-registry>/kai-controller:tag
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

### How it works
This project uses kubebuilder to scaffold the creation of kubernetes resources and controllers. Kai defines the following custom resources:
- App (high level application resource that encapsulates lower level resources, all other resources are configured through an App)
- Router (networking resource that controls routing and visibility of application versions)
- Config (contains application level configuration)
- Version (an immutable deployment that represents a specific config version)

#### Example app.yaml
An example http-echo app with a single route defined
```yaml
apiVersion: core.kai.io/v1alpha1
kind: App
metadata:
  name: http-echo-app
spec:
  template:
    spec:
      containers:
      - name: http-echo
        image: "hashicorp/http-echo"
        args: ["-listen=:9001", "-text=hello from version 1"]
        ports:
        - containerPort: 9001
  route:
    parentRefs:
    - name: kai-gateway
    rules:
    - matches:
      - path:
          type: PathPrefix
          value: /v1
      filters:
      - type: URLRewrite
        urlRewrite:
          path:
            type: ReplacePrefixMatch
            replacePrefixMatch: /
      backendRefs:
      - name: http-echo-app-00001-service
        port: 9001

```

### Project notes
This project draws inspiration from both KNative and KServe. It hopes to provide a modern and simple solution to the same problems those projects hope to solve.

Project Goals:
- Minimal dependencies
- UX
- Extensibility while relying on core resources (e.g.; ML model deployment)
- Make it easier to adopt Kubernetes and its best practices

### Project status
This project is currently in early alpha. Deployment to production is not recommended. As a project in early alpha you can expect frequent breaking and non-breaking changes.

We hope to gather community feedback around all aspects of the project both current and future.

## Contributing
If you would like to provide any feedback or have suggestions for new features please open an issue.

### For more info...
Please visit [dreamstax.io](https://dreamstax.io)
## License

// TODO
