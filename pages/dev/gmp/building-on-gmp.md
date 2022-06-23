#  Building applications with GMP

There are four simple steps (two mandatory and two optional):
1. Call the contract.
    - `callContract`, or
    - `callContractWithToken`.
2. Pay gas to the Gas Services contract.
3. (optional) Check the status of the call.
4. (optional) Execute and recovery methods.

### Step 1: Call the contract

Either:
1. Send messages using [[callContract](gmp-messages)], or
2. Send tokens with messages using [[callContractWithToken](gmp-tokens-with-messages)].

### Step 2: Pay gas to the Gas Services contract

See [[Gas Services](gas-services/overview)].

### Step 3: (optional) Check the status of the call

See [[Monitoring state of transactions](gmp-tracker-recovery/monitoring)].

### Step 4: (optional) Execute and recovery methods

Axelar network provides an optional relayer service, called Gas Services, which automatically relays transactions and executes approved messages. 

However, due to chains' conditions (either from the source or destination or both chains), some unexpected scenarios can occur, and the transaction might not be successfully executed. Axelar provides two options to recover these possible cases. See [[Transaction Recovery](gmp-tracker-recovery/recovery)].
