# Debugging your smart contract

The transaction at the destination contract can be reverted because some line in the function contains invalid logic. In this case, the next steps are to identify and fix the problem, and redeploy the smart contract.

Here are some tools that may help you to investigate and address misbehavior in the destination contract.

## Tenderly

[Tenderly](tenderly.co) is a powerful debugging tool for simulating smart contracts. It can be used for debugging the verified smart contract with a [breakpoint](https://en.wikipedia.org/wiki/Breakpoint). Tenderly also has other useful features for smart contract development, such as forking, transaction monitoring, etc. See their full documentation [here](https://docs.tenderly.co/).

Note that your contract must be verified at the block explorer or Tenderly (you can follow the guide [here](https://docs.tenderly.co/monitoring/verifying-a-smart-contract)). Also, the networks it supports are limited. [Here](https://docs.tenderly.co/supported-networks-and-languages) is a list of all the networks Tenderly supports.

Now, let's take a look at an example that will walk you through how you can use Tenderly to debug a failed transaction.

1. Consider this [failed transaction](https://testnet.axelarscan.io/gmp/0xeaaf091c0f435447c0a84e9d8cf1bc6f6ba3b7a2e5da277dbf37911fbc364d6e). If you click the `Execute` button to manually execute it from the wallet, Axelarscan will show a message indicating the transaction failed because there's something wrong with the destination contract's `executeWithToken` function.

![debugging-execution-reverted-01](/images/debugging-execution-reverted-01.png)

2. After logging in to the [Tenderly Dashboard](https://dashboard.tenderly.co/), select the `Contracts` menu from the sidebar, then click the `Add Contracts` button.

![debugging-execution-reverted-02](/images/debugging-execution-reverted-02.png)

3. Enter the destination-contract address by copying it from Axelarscan and click `Import Contracts`.

![debugging-execution-reverted-03](/images/debugging-execution-reverted-03.png)

4. Select the `Simulation` menu and click the `New Simulation` button.

![debugging-execution-reverted-04](/images/debugging-execution-reverted-04.png)

5. Fill the transaction info as shown below.

![debugging-execution-reverted-05](/images/debugging-execution-reverted-05.png)

6. You can customize transaction parameters if you want (optional). In this case, the default parameters are fine. If everything okays, then click the `Simulate Transaction` button.

![debugging-execution-reverted-06](/images/debugging-execution-reverted-06.png)

7. Scroll down to the bottom, you will see the last execution line in the smart contract code. Click the `View In Debugger` button to see more details.

![debugging-execution-reverted-07](/images/debugging-execution-reverted-07.png)

8. At this step, you can see the code that is causing an issue. The problem is that this contract passed `amount` instead of `sentAmount` into the `transfer` function, so it has insufficient funds to call a transfer function. That's the reason why this transaction was reverted.

![debugging-execution-reverted-08](/images/debugging-execution-reverted-08.png)
