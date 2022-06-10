# Steps in the creation of this project
1. Create the project directory and initialise a git repo there connecting to the remote repo
2. Run `operator-sdk init --domain beezlabs.app --repo github.com/beezlabs-org/cloudflare-tunnel-operator`
3. Run `operator-sdk create api --group cloudflare-tunnel-operator --version v1alpha1 --kind CloudflareTunnel --resource --controller`