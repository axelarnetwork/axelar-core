# AxelarJS SDK

The AxelarJS SDK is a `npm` dependency that empowers developers to make requests into the Axelar network from a frontend.

# Get started

- [Stable (v0.4.xx)](./sdk/axelarjs-stable)
- [Alpha (v0.5.xx)](./sdk/axelarjs-alpha)

## Overview

![Architecture diagram](/images/sdk-diagram.png)

Any request from the JS SDK is routed through a node REST server that redirects requests through a coordinated collection of microservices controlled by Axelar.

These microservices facilitate the relay of cross-chain transactions that run on top of the Axelar network.
