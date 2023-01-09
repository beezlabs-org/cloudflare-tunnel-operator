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
1. Figure out the service that you want to connect to. In the below example, the service looks as following
    ```yaml
    apiVersion: v1
    kind: Service
    metadata:
      name: traefik
      namespace: traefik
    spec:
      ports:
        - name: web
          nodePort: 30391
          port: 80
          protocol: TCP
          targetPort: web
        - name: websecure
          nodePort: 31118
          port: 443
          protocol: TCP
          targetPort: websecure
      selector:
        app.kubernetes.io/instance: traefik
        app.kubernetes.io/name: traefik
      type: LoadBalancer
    ```
2. Create an API token in cloudflare which has access to all Zones and the DNS.
3. Apply the examples updating the resources where needed
   ```sh
   kubectl apply -f examples/sampleTunnel/
   kubectl apply -f examples/secret/
   ```
4. Check CloudflareTunnel object
   ```console
   kubectl get cloudflaretunnel -A
   NAMESPACE    NAME            AGE
   cloudflare   sample-tunnel   3d21h
   ```
5. Access the URL that is set as domain
   ```bash
   https://example.sayakm.me
   ```

## Values

| Key                       | Type   | Default                                             | Description                            |
|---------------------------|--------|-----------------------------------------------------|----------------------------------------|
| image.repository          | string | `"ghcr.io/beezlabs-org/cloudflare-tunnel-operator"` | The image of the operator              |
| image.pullPolicy          | string | `"IfNotPresent`                                     | The image pull policy for the operator |
| image.tag                 | string | `"v0.1.0"`                                          | The image tag ofe the operator         |

Current values file [here](https://github.com/beezlabs-org/cloudflare-tunnel-operator/blob/main/charts/values.yaml)

```yaml
replicaCount: 1

image:
   repository: ghcr.io/beezlabs-org/cloudflare-tunnel-operator
   pullPolicy: IfNotPresent
   tag: 0.1.0

imagePullSecrets: []
nameOverride: ""
fullnameOverride: ""

serviceAccount:
   create: true
   annotations: {}
   name: ""

podAnnotations: {}

podSecurityContext: {}
# fsGroup: 2000

securityContext: {}
        # capabilities:
        #   drop:
        #   - ALL
        # readOnlyRootFilesystem: true
        # runAsNonRoot: true
# runAsUser: 1000

resources: {}
        # limits:
        #   cpu: 100m
        #   memory: 128Mi
        # requests:
        #   cpu: 100m
#   memory: 128Mi

autoscaling:
   enabled: false
   minReplicas: 1
   maxReplicas: 100
   targetCPUUtilizationPercentage: 80
   # targetMemoryUtilizationPercentage: 80

nodeSelector: {}

tolerations: []

affinity: {}
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
[app-version-shield]: https://shields.beezlabs.app/static/v1?label=App%20Version&message=v0.2.0&color=007ec6&style=for-the-badge
[app-version-url]: https://github.com/beezlabs-org/cloudflare-tunnel-operator/releases/latest
[chart-version-shield]: https://shields.beezlabs.app/static/v1?label=Chart%20Version&message=0.2.0&color=007ec6&style=for-the-badge
[chart-version-url]: https://github.com/beezlabs-org/cloudflare-tunnel-operator/releases/latest
