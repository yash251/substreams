In this guide, you'll learn how to initialize a Injective-based Substreams project within the Dev Container.

## Prerequisites

- Docker and VS Code installed and up-to-date.
- Visit the [Getting Started Guide](https://github.com/streamingfast/substreams-starter) to initialize your Dev Container.

## Step 1: Initialize Your Injective Substreams Project

1. Open the [Dev Container](https://github.com/streamingfast/substreams-starter) and follow the on-screen steps to initialize your project.
    
2. Running `substreams init` will give you the option to choose between two Injective project options. Select the one that best fits your requirements:
    - **Injective-minimal**: Creates a simple Substreams that extracts raw Injective block data and generates corresponding Rust code. This path will start you with the full raw block, you can navigate to the `substreams.yaml` (the manifest) to modify the input.
    - **Injective-events**: Creates a Substreams that extracts Injective events using the cached [Injective Foundational Module](https://substreams.dev/packages/injective-common/v0.2.4), filtered by one or more smart contract addresses. This includes type `wasm` events.

{% hint style="info" %} 
Tip: Have the start block of your transaction or specific events ready. 
{% endhint %}

## Step 2: Visualize the Data

1. Run `substreams auth` to create your [account](https://thegraph.market/) and generate an authentification token (JWT), then pass this token back as input.

2. Now you can freely use the `substreams gui` to visualize and itterate on your extracted data.

## Step 2.5: (Optionally) Transform the Data 

Within the generated directories, modify your Substreams modules to include additional filters, aggregations, and transformations, then update the manifest accordingly. To learn more about this, visit the [How-to-Guides](../../how-to-guides/develop-your-own-substreams/cosmos/injective/injective.md).

## Step 3: Load the Data

To make your Substreams queriable (as opposed to [direct streaming](../how-to-guides/sinks/stream/stream.md)), you can automatically generate a Subgraph (known as a [Substreams-powered subgraph](https://thegraph.com/docs/en/sps/introduction/)) or SQL-DB sink. 

## Additional Resources

You may find these additional resources helpful for developing your first Injective application.

### Dev Container Reference

The [Dev Container Reference](../references/devcontainer-ref.md) helps you navigate the container and its common errors. 

### CLI Reference

The [CLI reference](../references/cli/command-line-interface.md) lets you explore all the tools available in the Substreams CLI.

### Substreams Components Reference

The [Components Reference](../references/substreams-components/) dives deeeper into navigating the `substreams.yaml`.
