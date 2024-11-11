# Security Policy

## Introduction

Security researchers are essential in identifying vulnerabilities that may impact the Axelar Network. If you have discovered a security vulnerability in the Axelar chain or any repository managed by Axelar, we encourage you to notify us using one of the methods outlined below.

### Guidelines for Responsible Vulnerability Testing and Reporting

1. **Refrain from testing vulnerabilities on our publicly accessible environments**, including but not limited to:

    - Axelar mainnet
    - Axelar Powered Frontend Apps e.g satellite.money, Squid etc.
    - Axelar Testnet
    - Axelar Testnet Powered Frontend Apps e.g testnet.satellite.money

2. **Avoid reporting security vulnerabilities through public channels, including GitHub issues**

## Reporting Security Issues

To privately report a security vulnerability, please choose one of the following options:

### 1. Email

Send your detailed vulnerability report to `security@interoplabs.io`.

### 2. Bug Bounty Program

Axelar is partnered with Immunefi to offer a bug bounty program. Please visit [Immunefi's website](https://immunefi.com/bug-bounty/axelarnetwork/information/) for more information.

## Submit Vulnerability Report

When reporting a vulnerability through either method, please include the following details to aid in our assessment:

- Type of vulnerability
- Description of the vulnerability
- Steps to reproduce the issue
- Impact of the issue
- Explanation of how an attacker could exploit it

> [!NOTE]
> Review our criteria in the [Official Docs](https://docs.axelar.dev/resources/bug-bounty/#vulnerability-criteria)

## Vulnerability Disclosure Process

1. **Initial Report**: Submit the vulnerability via one of the above channels.
2. **Confirmation**: We will confirm receipt of your report within 48 hours.
3. **Assessment**: Our security team will evaluate the vulnerability and inform you of its severity and the estimated time frame for resolution.
4. **Resolution**: Once fixed, you will be contacted to verify the solution.
5. **Public Disclosure**: Details of the vulnerability may be publicly disclosed after approval from the team, ensuring it poses no further risk.

During the vulnerability disclosure process, we ask security researchers to keep vulnerabilities and communications around vulnerability submissions private and confidential until a patch is developed. Should a security issue require a network upgrade, additional time may be needed to raise a governance proposal and complete the upgrade.

During this time:

- Avoid exploiting any vulnerabilities you discover.
- Demonstrate good faith by not disrupting or degrading Axelar's services.

## Severity Characterization

| Severity     | Description                                                             |
|--------------|-------------------------------------------------------------------------|
| **CRITICAL** | Immediate threat to critical systems (e.g. funds at risk) |
| **HIGH**     | Significant impact on major functionality                               |
| **MEDIUM**   | Impacts minor features or exposes non-sensitive data                    |
| **LOW**      | Minimal impact                                                          |

## Bug Bounty

Our bug bounty program is managed by Immunefi. Please visit [Immunefi's website](https://immunefi.com/bug-bounty/axelarnetwork/information/) for more information.

> [!WARNING]
> Targeting our production environments will disqualify you from receiving any bounty.

## Feedback on this Policy

For recommendations on how to improve this policy, either submit a pull request or send an email to `security@axelar.network`.
