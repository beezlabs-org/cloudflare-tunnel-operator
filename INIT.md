# Steps in the creation of this project
1. Create the project directory and initialise a git repo there connecting to the remote repo
2. Run `kubebuilder init --domain beezlabs.app --repo github.com/beezlabs-org/cloudflare-tunnel-operator --owner "Beez Innovation Labs"`
3. Run `kubebuilder create api --group cloudflare-tunnel-operator --version v1alpha1 --kind CloudflareTunnel --resource --controller`
4. After changing anything in the API types, run `make manifests`
5. To run in Dev
   1. Run `make build` to generate the `manager` binary
   2. Run `make install` to upload the generated CRD
   3. Run the operator locally either using `make run` or using any IDE with the main command being `go run ./main.go`
