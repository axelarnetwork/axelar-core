# Debugging Your Smart Contract

The transaction at the destination contract can be reverted because some line in the function has an invalid logic. If it happens, the next step is to identify the problem, so we can know where to fix the issue and redeploy the smart contract.

So here are some useful tools that will help you to investigate misbehavior in the destination contract.

## Tenderly

[Tenderly](tenderly.co) is a powerful debugging tool and simulating smart contracts. It can be used for debugging the verified smart contract with a [breakpoint](https://en.wikipedia.org/wiki/Breakpoint). Tenderly also has other useful features for smart contract development such as forking, transaction monitoring, etc. See their full documentation [here](https://docs.tenderly.co/).

Note that your contract must be verified at the block explorer or Tenderly (you can follow the guide [here](https://docs.tenderly.co/monitoring/verifying-a-smart-contract)). Also, the networks it supported are limited. Here're all supported networks.

![debugging-tenderly-supported-network](/images/debugging-tenderly-supported-network.png)

Now, let's take a look at an example that will walk you through on how you can use it to debug a failed transaction.

1. Considering this [failed transaction](https://testnet.axelarscan.io/gmp/0xeaaf091c0f435447c0a84e9d8cf1bc6f6ba3b7a2e5da277dbf37911fbc364d6e). If we click the `Execute` button to manually execute it from our wallet, The Axelarscan will show a message indicating it is failed because there's something wrong with the destination contract's `executeWithToken` function.

![debugging-execution-reverted-01](/images/debugging-execution-reverted-01.png)

2. After logged-in to the [Tenderly Dashboard](https://dashboard.tenderly.co/), click at the `Contracts` menu from the sidebar, then click `Add Contracts` button.

![debugging-execution-reverted-02](/images/debugging-execution-reverted-02.png)

3. Then, enter the destination contract address by copying it from the Axelarscan and clicks "Import Contracts".

![debugging-execution-reverted-03](/images/debugging-execution-reverted-03.png)

4. Then, clicks the Simulation menu and clicks at a "New Simulation" button.

![debugging-execution-reverted-04](/images/debugging-execution-reverted-04.png)

5. Then, fill the transaction infos like below.

![debugging-execution-reverted-05](/images/debugging-execution-reverted-05.png)

6. You can customize transaction parameters if you want (Optional). But in this case, we can use the default parameters. If everything okays, then clicks a "Simulate Transaction" button.

![debugging-execution-reverted-06](/images/debugging-execution-reverted-06.png)

7. Scroll down to the buttom, you will see the last execution line in the smart contract code. Clicks a "View In Debugger" button to see more details.

![debugging-execution-reverted-07](/images/debugging-execution-reverted-07.png)

8. At this step, we can see the code that causes an issue. The problem is that this contract passed `amount` instead of `sentAmount` into the `transfer` function, so it has insufficient funds to call a transfer function. That's the reason why this transaction was reverted.

![debugging-execution-reverted-08](/images/debugging-execution-reverted-08.png)
