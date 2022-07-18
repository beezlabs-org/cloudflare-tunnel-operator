[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]
[![License][license-shield]][license-url]
[![Artifact Hub][artifact-hub-shield]][artifact-hub-url]
[![Go][go-shield]][go-url]
[![App Version][app-version-shield]][app-version-url]
[![Chart Version][chart-version-shield]][chart-version-url]
<br />

<h1 align="center">Cloudflare Tunnel Operator</h1>

<p align="center">
  Kubernetes operator to run Cloudflare Tunnels
  <br />
  <a href="https://github.com/beezlabs-org/cloudflare-tunnel-operator"><strong>Explore the docs »</strong></a>
  <br />
  <br />
  <a href="https://github.com/beezlabs-org/cloudflare-tunnel-operator/issues">Report Bug</a>
  ·
  <a href="https://github.com/beezlabs-org/cloudflare-tunnel-operator/issues">Request Feature</a>
</p>

## About this project

[Cloudflare Tunnels](https://www.cloudflare.com/en-gb/products/tunnel/) allows us to access systems behind a firewall
or ones without a static IP among other things. This can be incredibly useful for people running their own home servers
as it enables them to expose applications running on those servers to the internet without needing to port forward
or pay for a static IP.

But what about those running home Kubernetes clusters? One could run the tunnel application like any other application
on Kubernetes but creating and destroying tunnels would either involve changing the config file in the cluster or
using the Cloudflare Tunnel dashboard. This is certainly a valid way but might not be the ideal way, especially if you
live and breathe [GitOps](https://www.weave.works/technologies/gitops/).

That is where this project comes in. It allows you to create custom resources for Cloudflare Tunnels in your Kubernetes
cluster. This means that the Tunnel will be active as long as your custom resource exists and will be updated if your
custom resource is updated.

### Built With

- [Operator SDK](https://sdk.operatorframework.io/)
- [Kubebuilder](https://book.kubebuilder.io/)
- [cloudflared](https://github.com/cloudflare/cloudflared/)
- [cloudflare-go](https://github.com/cloudflare/cloudflare-go/)

## Installing

To install the chart with the release name `my-cloudflare-tunnel-operator`:

```sh
helm repo add beezlabs https://charts.beezlabs.app/
helm install my-cloudflare-tunnel-operator beezlabs/cloudflare-tunnel-operator --version 0.1.0
```

## Usage
1. Apply some example
```sh
kubectl apply -f examples/deployment/
kubectl apply -f examples/http/simple/
kubectl apply -f examples/http/full-configuration/
```
2. Check ngrok object
```console
kubectl get ngrok --all-namespaces
NAMESPACE    NAME                       STATUS    URL
default      http-simple                created   https://9496e56ed0bc.ngrok.io
default      http-full-configuration    created   https://ngrok.zufardhiyaulhaq.com
```

3. access the URL
```console
https://d5150f7c3588.ngrok.io
https://ngrok.zufardhiyaulhaq.com
```

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| operator.image | string | `"zufardhiyaulhaq/ngrok-operator"` |  |
| operator.replica | int | `1` |  |
| operator.tag | string | `"v1.3.0"` |  |
| resources.limits.cpu | string | `"200m"` |  |
| resources.limits.memory | string | `"100Mi"` |  |
| resources.requests.cpu | string | `"100m"` |  |
| resources.requests.memory | string | `"20Mi"` |  |

see example files [here](https://github.com/zufardhiyaulhaq/ngrok-operator/blob/master/charts/ngrok-operator/values.yaml)

```yaml
operator:
  # image of ngrok-operator
  image: "zufardhiyaulhaq/ngrok-operator"
  # tag of ngrok-operator image
  tag: "v1.3.0"
  # number of replica for deployment
  replica: 1

resources:
  limits:
    cpu: 200m
    memory: 100Mi
  requests:
    cpu: 100m
    memory: 20Mi
```

## License

Licensed under the Apache License, Version 2.0. Copyright Beez Innovation Labs Pvt Ltd.

## Contact

Sayak Mukhopadhyay - [SayakMukhopadhyay](https://github.com/SayakMukhopadhyay) - sayak@beezlabs.com


<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->

[contributors-shield]: https://shields.beezlabs.app/github/contributors/beezlabs-org/cloudflare-tunnel-operator?style=for-the-badge
[contributors-url]: https://github.com/beezlabs-org/cloudflare-tunnel-operator
[forks-shield]: https://shields.beezlabs.app/github/forks/beezlabs-org/cloudflare-tunnel-operator?style=for-the-badge
[forks-url]: https://github.com/beezlabs-org/cloudflare-tunnel-operator/network/members
[stars-shield]: https://shields.beezlabs.app/github/stars/beezlabs-org/cloudflare-tunnel-operator?style=for-the-badge
[stars-url]: https://github.com/beezlabs-org/cloudflare-tunnel-operator/stargazers
[issues-shield]: https://shields.beezlabs.app/github/issues/beezlabs-org/cloudflare-tunnel-operator?style=for-the-badge
[issues-url]: https://github.com/beezlabs-org/cloudflare-tunnel-operator/issues
[license-shield]: https://shields.beezlabs.app/github/license/beezlabs-org/cloudflare-tunnel-operator?style=for-the-badge
[license-url]: https://github.com/beezlabs-org/cloudflare-tunnel-operator/blob/master/LICENSE
[artifact-hub-shield]: https://shields.beezlabs.app/static/v1?label=Artifact%20Hub&message=beezlabs&color=417598&logo=artifacthub&style=for-the-badge
[artifact-hub-url]: https://artifacthub.io/packages/helm/beezlabs/cloudflare-tunnel-operator
[go-shield]: https://shields.beezlabs.app/static/v1?label=Go&message=v1.17&color=007ec6&style=for-the-badge
[go-url]: https://go.dev/
[app-version-shield]: https://shields.beezlabs.app/static/v1?label=App%20Version&message=v{{ .Env.APP_VERSION }}&color=007ec6&style=for-the-badge
[app-version-url]: https://github.com/beezlabs-org/cloudflare-tunnel-operator/releases/latest
[chart-version-shield]: https://shields.beezlabs.app/static/v1?label=Chart%20Version&message={{ .Env.CHART_VERSION }}&color=007ec6&style=for-the-badge
[chart-version-url]: https://github.com/beezlabs-org/cloudflare-tunnel-operator/releases/latest
