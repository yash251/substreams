# Getting Started with Substreams

Integrating Substreams can be quick and easy. This guide will help you get started with consuming ready-made Substreams packages or developing your own. Substreams are permissionaless. Grab a key [here](thegraph.market), no personal information required, and start streaming on-chain data.

# Build

## Explore Available Substreams Packages

There are many ready-to-use Substreams packages available. You can explore these packages using the [**Substreams Registry**](https://substreams.dev). The registry lets you search for and find packages that meet your needs.

Once you find a package that fits your needs, you can choose how you want to consume the data:
- **SQL Database**: Send the data to a database.
- **Subgraph**: Configure an API to meet your data needs and host it on The Graph Network.
- **Direct Streaming**: Stream data directly from your application.

<figure><img src=".gitbook/assets/intro/consume-flow.png" width="100%" /></figure>

## Optionally Develop Your Own Substreams

If you can't find a Substreams package that meets your specific needs, you can develop your own. Substreams are built with Rust, so you’ll write functions that extract and filter the data you need from the blockchain. The easiest way to get started is by referring to the ecosystem specific tutorial, enabling you to quickly filter data: 

- [EVM](./tutorials/evm.md)
- [Solana](./tutorials/solana.md)
- [Starknet](./tutorials/starknet.md)
- [Injective](./tutorials/cosmos-compatible/injective.md)
- [Mantra](./tutorials/cosmos-compatible/mantra.md)

To build and optimize your Substreams from zero, use the minimal path within the [Dev Container](./references/devcontainer-ref.md) to setup your environment and follow the [How-To Guides](./how-to-guides/develop-your-own-substreams/develop-your-own-substreams.md).

## Learn

- **Substreams Architecture:**:  For a deeper understanding of how Substreams works, explore the [architectural overview](architecture.md) of the data service.
- **Substreams Reliability Guarantees**: With a simple reconnection policy, Substreams guarantees you'll never miss data [Reliability Guarantees](./references/reliability-guarantees.md).
- **Supported Networks**: Check-out which endpoints are supported [here](./references/chains-and-endpoints.md).