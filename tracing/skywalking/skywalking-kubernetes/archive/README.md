Apache SkyWalking Kubernetes
==========

# Deploy SkyWalking backend to Kubernetes cluster

To install and configure skywalking in a Kubernetes cluster, follow these instructions.

## Prerequisites

Please promise the `skywalking` namespace existed in the cluster, otherwise, create a new one.

`kubectl apply -f namespace.yml`

## Deploy Elasticsearch

Use `kubectl apply -f ` with the scripts in `elasticsearch` to deploy elasticsearch servers
in the cluster.

> `01-storageclass.yml` assume to use GKE as the kubernetes provisioner. You could fix it according
to your kubernetes environment.

## Deploy OAP server 

Use `kubectl apply -f ` with the scripts in `oap` to deploy oap server
in the cluster.

## Deploy UI server 

Use `kubectl apply -f ` with the scripts in `ui` to deploy oap server
in the cluster.

# Setup Istio to send metric to oap

## Prerequisites

Istio should be installed in kubernetes cluster.

## Setup Istio to send metric to oap

Use `kubectl apply -f ` with the scripts in `istio` to setup.
