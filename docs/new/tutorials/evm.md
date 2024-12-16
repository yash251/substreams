In this guide, you'll learn how to initialize an EVM-based Substreams project within the Dev Container.

## Step 1: Initialize Your EVM Substreams Project

1. Open the [Dev Container](https://github.com/streamingfast/substreams-starter) and follow the on-screen steps or `README.md` to initialize your project.
    
2. Running `substreams init` will give you the option to choose between two EVM project options. Select the one that best fits your requirements:
    - **evm-minimal**: Creates a simple Substreams that extracts raw EVM block data and generates corresponding Rust code. This path will start you with the full raw block, you can navigate to the `substreams.yaml` (the manifest) to modify the input.
    - **evm-events-calls**: Creates a Substreams that extracts and decodes EVM events and calls using the cached [EVM Foundational Module](https://substreams.dev/streamingfast/ethereum-common/v0.3.0), filtered by one or more smart contract addresses. Contract ABIs are retrieved from Etherscan. If an ABI isn’t available, you’ll need to provide it yourself.

## Step 2: Visualize the Data

1. Running `substreams auth` will prompt you to create your account [here](https://thegraph.market/) to generate an authentification token (JWT), pass it back as input.

2. Now you can freely use the `substreams gui` to visualize and itterate on your extracted data.

## Step 2.5: (Optionally) Transform the Data 

Within the generated directories, modify your Substreams modules to include additional filters, aggregations, and transformations, then update the manifest accordingly. To learn more about this, visit the [How-to-Guides](../how-to-guides/develop-your-own-substreams/evm/exploring-ethereum/exploring-ethereum.md)

## Step 3: Load the Data

To make your Substreams queriable (as opposed to [direct streaming](../how-to-guides/sinks/stream/stream.md)), you can automatically generate a Subgraph (known as a [Substreams-powered subgraph](https://thegraph.com/docs/en/sps/introduction/)) or SQL-DB sink by following the on-screen steps or referring to the `README.md`. 

## Additional Resources

You may find these additional resources helpful for developing your first EVM application.

### Dev Container Reference

The [Dev Container Reference](../references/devcontainer-ref.md) helps you navigate the container and its common errors. 

### CLI Reference

The [CLI reference](../references/cli/command-line-interface.md) lets you explore all the tools available in the Substreams CLI.

### Substreams Components Reference

The [Components Reference](../references/substreams-components/) dives deeeper into navigating the `substreams.yaml`.

