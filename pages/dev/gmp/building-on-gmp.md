#  Building applications with GMP

There are four simple steps (two mandatory and two optional):
1. Call the contract.
    - `callContract`, or
    - `callContractWithToken`.
2. Pay gas to the Gas Services contract.
3. *(optional)* Check the status of the call.
4. *(optional)* Execute and recover transactions.

### Step 1: Call the contract

Either:
1. Send messages using [[callContract](gmp-messages)], or
2. Send tokens with messages using [[callContractWithToken](gmp-tokens-with-messages)].

### Step 2: Pay gas to the Gas Services contract

See [[Gas Services](gas-services/overview)].

### Step 3: Check the status of the call *(optional)* 

See [[Monitoring state of transactions](gmp-tracker-recovery/monitoring)].

### Step 4: Execute and recover transactions *(optional)*

Axelar network provides an optional relayer service, called Executor Service, which automatically relays transactions and executes approved messages. See [[Executor Service](/dev/gmp/executor-service)].

Sometimes, unexpected scenarios occur due to chains' conditions (either from the source or destination or both chains) which cause unsuccessful executed transactions. Axelar provides options to recover them. See [[Transaction Recovery](gmp-tracker-recovery/recovery)].