In this guide, you'll learn how to initialize a Solana-based Substreams project within the Dev Container.

## Step 1: Initialize Your Solana Substreams Project

1. Open the [Dev Container](https://github.com/streamingfast/substreams-starter) and follow the on-screen steps or `README.md` to initialize your project.

2. Running `substreams init` will give you the option to choose between two Solana project options. Select the one that best fits your requirements:
    - **sol-minimal**: Creates a simple Substreams that extracts raw Solana block data and generates corresponding Rust code. This path will start you with the full raw block, you can navigate to the `substreams.yaml` (the manifest) to modify the input.
    - **sol-transactions**: Creates a Substreams that filters Solana transactions based on one or more Program IDs and/or Account IDs, using the cached [Solana Foundational Module](https://substreams.dev/streamingfast/solana-common/v0.3.0).
    - **sol-anchor-beta**: Given an Anchor IDL, create a Substreams that decodes instructions and events. If an IDL isn’t available using the `idl` subcommand within the [Anchor ClI](https://www.anchor-lang.com/docs/cli), you’ll need to provide it yourself.

{% hint style="info" %} 
Note: The filtered_transactions_without_votes module extracts transactions while excluding voting transactions, reducing data size and costs by 75%. To access voting transactions, use a full Solana block.
{% endhint %}
    
## Step 2: Visualize the Data

1. Running `substreams auth` will prompt you to create your account [here](https://thegraph.market/) to generate an authentification token (JWT), pass it back as input.

2. Now you can freely use the `substreams gui` to visualize and itterate on your extracted data.

## Step 2.5: (Optionally) Transform the Data 

Within the generated directories, modify your Substreams modules to include additional filters, aggregations, and transformations, then update the manifest accordingly. To learn more about this, visit the [How-to-Guides](../how-to-guides/develop-your-own-substreams/solana/solana.md)

## Step 3: Load the Data

To make your Substreams queriable (as opposed to [direct streaming](../how-to-guides/sinks/stream/stream.md)), you can automatically generate a Subgraph (known as a [Substreams-powered subgraph](https://thegraph.com/docs/en/sps/introduction/)) or SQL-DB sink by following the on-screen steps or referring to the `README.md`. 

## Additional Resources

You may find these additional resources helpful for developing your first Solana application.

### Dev Container Reference

The [Dev Container Reference](../references/devcontainer-ref.md) helps you navigate the container and its common errors. 

### CLI Reference

The [CLI reference](../references/cli/command-line-interface.md) lets you explore all the tools available in the Substreams CLI.

### Substreams Components Reference

The [Components Reference](../references/substreams-components/) dives deeeper into navigating the `substreams.yaml`.

