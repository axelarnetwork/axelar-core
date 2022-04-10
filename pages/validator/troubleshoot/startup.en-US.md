# Start-up issues

import Callout from 'nextra-theme-docs/callout'

[TODO revise]

If the process was missing, check if `tofnd` is running. Install the `nmap` command if you do not have it, and check the tofnd port

```bash
nmap -p 50051 localhost
```

Look for the `STATE` of the port, which should be `open` or `closed`. If the port is `closed`, restart your node and ensure tofnd is running. If the port is `open`, then there is a connection issue between vald and tofnd.

To fix the connectivity issue, find the `tofnd` container address manually and provide it to `vald`.
Find the `tofnd` address.

```bash
docker inspect tofnd
```

Near the bottom of the JSON output, look for `Networks`, then `bridge`, `IPAddress`, and copy the address listed.
Next, ping the IP Address from inside `Axelar Core` to see if it works. Install the `ping` command if it does not exist already.

```bash
docker exec axelar-core ping {your tofnd IP Address}
```

eg:

```bash
docker exec axelar-core ping 172.17.0.2
```

You should see entries starting to appear one by one if the connection succeeded. Stop the ping with `Control + C`.
Save this IP address.

Next, query your validator address with

```bash
docker exec axelar-core axelard keys show validator --bech val -a
```

<Callout emoji="📝">
  Note: Verify that the returned validator address starts with `axelarvaloper`
</Callout>

Now, start `vald`, providing the IP address and validator address:

```bash
docker exec axelar-core axelard vald-start --tofnd-host {your tofnd IP Address} --validator-addr {your validator address} --node {your axelar-core IP address}
```
eg:
```bash
docker exec axelar-core axelard vald-start --tofnd-host 172.17.0.2 --validator-addr axelarvaloper1y4vplrpdaqplje8q4p4j32t3cqqmea9830umwl
```



Your vald should be connected properly. Confirm this by running the following and looking for an `vald-start` entry.
```bash
docker exec axelar-core ps
```