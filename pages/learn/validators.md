# Registering External Chains for Validators

Learn about what happens when Axelar network adds support for new chains, and the role that Axelar validators play in enabling these new chains.

## Background

Axelar provides cross-chain interoperability across all supported chains. These supported chains currently belong under two broad categories, being either EVM based chains, or Cosmos IBC based chains.The next section describes what role validators play when Axelar onboards a new chain. 

When a new chain is onboarded into Axelar, it can access and communicate with all the other Axelar connected chains right away. This is possible because Axelar acts as a universal overlay network, similar to a “hub and spoke” model. Since Axelar is a blockchain network itself, it is able to use routing logic to act as the interoperability network’s hub. This architecture means it is fairly easy to onboard additional chains to the network, especially when compared to a “pairwise bridge model”, where adding a single new chain requires connections to be formed between each existing chain and the new one.

Additionally, any validator is able to propose the addition of a new chain into the network. Depending on whether the chain is a Cosmos-based chain or an EVM-based chain, the specific steps will differ, as described later on. When a new chain is added into the network, it is important to have as many validators supporting the new chain as possible, but it is not required for every validator to support every chain. With many validators supporting a chain, the network becomes more decentralized, as even if one validator goes down the connection to the chain will still remain open and messages will continue to be relayed, if a large number of validators supporting the chain remain operational. To ensure each connected chain gets as much validator support as possible, Axelar validators are incentivized with additional network rewards for each connected chain they support.

## Role of validators when onboarding Cosmos IBC chains

Axelar network is built on top of the Cosmos SDK, meaning Axelar itself is a Cosmos-based chain. Within the Cosmos ecosystem, the Inter-Blockchain Communication (IBC) protocol allows cross-chain communication. Axelar uses IBC to onboard new Cosmos chains to the Axelar network, and give the onboarded chain access to all other Axelar connected chains, both Cosmos and EVM. Read the Cosmos SDK [IBC docs](https://ibc.cosmos.network/) to learn more about the protocol.

IBC is a permissionless protocol, meaning any Axelar validator can start the onboarding process for a new Cosmos chain, by setting up IBC channels between Axelar and the new chain. After a channel is established, Axelar validators are encouraged to support the channel. Axelar validators must run light clients between Axelar and the Cosmos chain, for each external cosmos chain they support. See the Axelar IBC info docs for more details. Having many validators supporting an IBC channel ensures decentralization. If one validator goes down, the IBC channel remains open and messages can still be relayed. 

Once an IBC channel is established, connecting the new chain to Axelar network with strong validator support, the Axelar network can register a path, connecting the IBC chain with all other chains supported by Axelar, both IBC and EVM. Incoming messages from the Cosmos chain to Axelar will be sent through the IBC channel, then processed and approved by Axelar validators, and executed on the destination chain. Outgoing messages from Axelar to the Cosmos chain will first be approved by Axelar validators, then executed on the Cosmos chain via the IBC channel.

## Role of validators when onboarding EVM chains

Cosmos chains benefit from having broad validator support; for EVM chains, it is required. Since inherently interoperable protocols like IBC do not exist between Axelar and EVM chains, Axelar support and security for EVM chains depends on the Axelar validators' participation by voting on EVM chain events in a proof-of-stake consensus model. When processing cross-chain messages from a source EVM chain, Axelar validators must query their RPC endpoints from their EVM chain to verify the submitted message. This means that in order to support a new EVM chain, Axelar validators must run a node for the new EVM chain.

The Axelar validator must obtain a secured and private RPC endpoint for the EVM chain. This is usually done by the validator running their own node for the EVM chain, and using their own RPC endpoint. After this is setup, the validator needs to inform Axelar about their support for this new chain, and submit a command to Axelar, registering their validator as a chain maintainer for the new EVM chain. If the amount of validators registering support for the new chain reaches a certain threshold, the EVM chain can be activated on the Axelar network layer, creating paths to and from the new chain and connecting it with every other Axelar supported chain, both IBC and EVM.

Note that it is not required for all Axelar validators to support the new EVM chain. As long as a set threshold of validators register as a chain maintainer, the ability to enable the chain at the network level is supported. However, additional rewards are given out to validators for each additional EVM chain they support, which encourages validators to support as many chains as they can.



