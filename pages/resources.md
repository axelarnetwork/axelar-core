# Resources

import Tabs from '../components/tabs'
import Resources from '../components/resources'

<div className="-mt-4">
  <Tabs
    tabs={[
      {
        title: "Mainnet",
        content: (
          <Resources
            environment="mainnet"
          />
        )
      },
      {
        title: "Testnet",
        content: (
          <Resources
            environment="testnet"
          />
        )
      },
      {
        title: "Testnet-2",
        content: (
          <Resources
            environment="testnet-2"
          />
        )
      }
    ]}
  />
</div>

## Looking for help?

Join the [Axelar discord](https://discord.gg/aRZ3Ra6f7D) and visit channels: [developers](https://discord.com/channels/770814806105128977/955655587260170272) | [testnet](https://discord.com/channels/770814806105128977/799299951078408242) | [general](https://discord.com/channels/770814806105128977/770814806105128980)
