# Kai
No frills pipelines

## Description
Kai automates the management and integration of [kai-piper](https://github.com/dreamstax/kai-piper) which enables users to easily define and execute pipelines. Kai also defines several higher level resources on top of the Kubernetes API that make it easier to define, deploy and manage your pipelines.

For more detailed information please see the docs at [kai-docs](https://kai-docs.dreamstax.io) and we encourage everyone to join the [kai-discord](https://discord.gg/qX4umFFkza) server.

## Table of Contents
- [Kai](#kai)
  - [Description](#description)
  - [Table of Contents](#table-of-contents)
  - [Installation](#installation)
  - [Usage](#usage)
  - [Features](#features)
  - [Contributing](#contributing)
  - [Acknowledgements](#acknowledgements)
  - [Project status](#project-status)
  - [License](#license)

## Installation
The quickest way to get up and running with Kai is to clone the repo and run the quickstart using make. This will create a cluster using [kind](https://sigs.k8s.io/kind), build and deploy the controller, and deploy an example pipeline.
```bash
git clone https://github.com/dreamstax/kai && cd kai
make quickstart
```
#### For Existing Clusters
*note: this section is wip as we do not yet have a first release*
Use `kubectl` to install the CRDs and deploy the controller to your cluster.
```bash
kubectl apply -f {github-release-url}
```
## Usage
Create a new pipeline by defining a Pipeline resource.
```yaml
# pipeline.yaml
apiVersion: core.kai.io/v1alpha1
kind: Pipeline
metadata:
  name: image-classifier
spec:
  steps:
  - spec:
      model:
        modelFormat: pytorch
        uri: gs://kfserving-examples/models/torchserve/image_classifier/v1
```

Then apply this pipeline resource to the cluster.
```bash
kubectl apply -f pipeline.yaml
```

#### Running a pipeline
*note: this section is wip as we build out kai-piper*
- retrieve pipeline ID (pipeline resource could expose this, also available via kai-piper)
- port-forward kai-piper server (could also be registered on an ingress gateway)
- call `/v1alpha1/pipelineJobs/{job_id}:run`

## Features
The Kai controller provides the following features
- Pipeline orchestration and management via [kai-piper](https://github.com/dreamstax/kai-piper)
- Service Deployment - Kai can register your pipelines and deploy your steps via a consolidated API (similar to KServe/KNative) which can make adopting Kubernetes a much easier process.
- Kubernetes native scaling - piper is a simple orchestrator and your steps run as standard Kubernetes deployments allowing you to configure HPA's and resources separately from the pipeline orchestration.
- Flexibility - Since steps are just deployments with services you can configure Kai to execute steps outside of it's control plane. This makes it easy for users with existing model deployments or workloads to migrate. Additionally since Kai is only orchestrating the execution of your steps your workloads won't have a large dependency on it's API.

## Contributing
For people interested in contributing to Kai please check out our open issues on GitHub. For bug reports or feature requests please add the appropriate tag to your issue so it can be triaged appropriately. We also strongly encourage members of the community to join the [kai-discord](https://discord.gg/qX4umFFkza) server, where you'll be able to communicate with maintainers and other members about Kai.

#### Prerequisites
In order to develop for Kai you'll need the following dependencies
- Go >= 1.20

#### Getting Started
The easiest way to get started is by cloning this repo and running make to install and setup all other dependencies

```sh 
git clone https://github.com/dreamstax/kai && cd kai
make quickstart
```
This is an exhaustive build and install of all required dependencies and assumes nothing exists. This is mostly useful for initial repo pulls or starting from scratch. You can see a list of commands that are run within the makefile and perform the ones necessary during development.

After completion of the make command, ensure the controller is running in the cluster by checking the pods within the namespace.
```sh
kubectl get pods -n kai-system
```

#### Running locally
Oftentimes when making changes to the controller it's nice to just run the controller locally and not in the cluster. To achieve this run the following command.

```sh
make dev
```
This will build the manifests and install the CRDs into the cluster then run the controller in your current terminal window. For a tighter iteration loop, familiarize yourself with the make targets and just run what you need.

## Acknowledgements
Kai aims to be a simple and modern solution for defining and deploying pipelines for local, research, or production environments.

This project draws inspiration from [KNative](https://github.com/knative/serving), [KServe](https://github.com/kserve/kserve), [Argo](https://github.com/argoproj/argo-workflows), and [Cadence](https://github.com/uber/cadence). We hope to provide a modern and simple alternative to some of the same problems those projects aim to solve.

## Project status
This project is currently in early alpha. Deployment to production is not recommended. As a project in early alpha you can expect frequent breaking and non-breaking changes.

We hope to gather community feedback around all aspects of the project both current and future.

## License
Copyright 2023 The Kai Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
