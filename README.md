# provider-zpa

Crossplane provider for [Zscaler ZPA]
The provider built from this repository can be installed into a Crossplane control plane or run seperately. It provides the following features:

* Extension of the K8s API with CRDs to represent zscaler zpa objects as K8s resources
* Controllers to provision these resources into a zscaler zpa instance
* Implementations of Crossplane's portable resource abstractions, enabling zscaler zpa resources to fulfill a user's general need for cloud services

## Getting Started and Documentation

For getting start with Crossplane setup installation and deployment, see the [official documentation](https://crossplane.io/docs/latest).

### To use provider-zpa

1. Create a new Zscaler ZPA API token, and store it in a K8s secret
```bash
kubectl create secret generic provider-zpa-creds --from-literal=token=$API_TOKEN -n crossplane-system
```
2. Create a new [ProviderConfig](examples/config/example-provider-config.yaml) resource with a references to this secret

You are now ready to create resources as described in [examples](examples).

## Contributing

provider-zpa is a community driven project and we welcome contributions. See the
Crossplane
[Contributing](https://github.com/crossplane/crossplane/blob/master/CONTRIBUTING.md)
guidelines to get started.

### Adding New Resource

New resources can be added by defining the required types in `apis` and the controllers `pkg/controllers/`.

To generate the CRD YAML files run

    make generate


## Report a Bug

For filing bugs, suggesting improvements, or requesting new features, please
open an [issue](https://github.com/haarchri/provider-zpa/issues).

## Contact

Please use the following to reach members of the community:

* Slack: Join our [slack channel](https://slack.crossplane.io)
* Forums:
  [crossplane-dev](https://groups.google.com/forum/#!forum/crossplane-dev)
* Twitter: [@crossplane_io](https://twitter.com/crossplane_io)
* Email: [info@crossplane.io](mailto:info@crossplane.io)

## Governance and Owners

provider-aws is run according to the same
[Governance](https://github.com/crossplane/crossplane/blob/master/GOVERNANCE.md)
and [Ownership](https://github.com/crossplane/crossplane/blob/master/OWNERS.md)
structure as the core Crossplane project.

## Code of Conduct

provider-zpa adheres to the same [Code of
Conduct](https://github.com/crossplane/crossplane/blob/master/CODE_OF_CONDUCT.md)
as the core Crossplane project.

## Licensing

provider-zpa is under the Apache 2.0 license.


## Usage

To run the project

    make run

To run all tests:

    make test

To build the project

    make build

To list all available options

    make help

[See more](./INSTALL.md)

## Code generation

See [CODE_GENERATION.md](./CODE_GENERATION.md)
