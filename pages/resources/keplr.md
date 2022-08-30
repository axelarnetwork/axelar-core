# Add a custom network to Keplr
import Button from '../../components/keplr/addKeplrWallet'
import ButtonLink from '../../components/button'
import AddKeplr from '../../components/keplr'
import Form from '../../components/textarea'

## Don't have the Keplr Chrome extension? Add it here

<ButtonLink buttonTitle="Add Keplr Wallet" url="https://chrome.google.com/webstore/detail/keplr/dmkamcknogkgcdfhhbddcghachkejeap?hl=en"></ButtonLink>

## Add Axelar-supported network

### Add Network to Keplr

Chain Name  | Testnet                                               | Mainnet                                               |
----------  | ----------------------------------------------------- | ----------------------------------------------------- |
Axelar      | <AddKeplr environment="testnet" chain="axelar" />     | Already Supported                                     |
Crescent    | <AddKeplr environment="testnet" chain="crescent" />   | <AddKeplr environment="mainnet" chain="crescent" />   |
e-Money     | N/A                                                   | Already Supported                                     |
Evmos       | <AddKeplr environment="testnet" chain="evmos" />      | N/A                                                   |
Fetch       | <AddKeplr environment="testnet" chain="fetch" />      | N/A                                                   |
Injective   | N/A                                                   | <AddKeplr environment="mainnet" chain="injective" />  |
Juno        | N/A                                                   | Already Supported                                     |
Kujira      | <AddKeplr environment="testnet" chain="kujira" />     | <AddKeplr environment="mainnet" chain="kujira" />     |
Osmosis     | <AddKeplr environment="testnet" chain="osmosis" />    | Already Supported                                     |
Secret      | N/A                                                   | Already Supported                                     |
Sei         | <AddKeplr environment="testnet" chain="axelar" />     | Already Supported                                     |
Terra-2     | <AddKeplr environment="testnet" chain="terra" />      | <AddKeplr environment="mainnet" chain="terra" />      |

## Add your custom network

<Form />

<Button buttonTitle="Validate Input & Add to Keplr" onClick={async (data) => {
    let field = document.getElementById("keplr_wallet_json_configuration");
    let settings = null;

    try {
        const ugly = field.value;
        document.getElementById('keplr_wallet_json_configuration').value = JSON.stringify(JSON.parse(ugly), undefined, 4);
    } catch (error) {
        if (error instanceof SyntaxError) {
            alert("There was a syntax error. Please correct it and try again: " + error.message);
        }
        else {
            throw error;
        }
    }

    try {
        settings = JSON.parse(field.value);

        try {
            await window.keplr.enable(settings.chainId)
            alert(settings.chainId + " already added")
        } catch (e) {
            console.log("Unable to connect to wallet natively, so trying experimental chain")
            try {
                await window.keplr.experimentalSuggestChain(settings)
                await window.keplr.enable(settings.chainId)
            } catch (e2) {
                alert("Invalid Keplr JSON configuration. " + e2.message);
                console.log("and yet there is a problem in trying to do that too", e2)
            }
        }

    } catch (error) {
        if (error instanceof SyntaxError) {
            alert("There was a syntax error. Please correct it and try again: " + error.message);
        }
        else {
            alert(error.message);
        }
    }
}}></Button>
