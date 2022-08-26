import { useState } from "react";
const placeholder = '{"rpc":"https://axelartest-rpc.quickapi.com:443","rest":"https://axelartest-lcd.quickapi.com:443","chainId":"axelar-testnet-lisbon-3","chainName":"Axelar Testnet","stakeCurrency":{"coinDenom":"AXL","coinMinimalDenom":"uaxl","coinDecimals":6},"bech32Config":{"bech32PrefixAccAddr":"axelar","bech32PrefixAccPub":"axelarpub","bech32PrefixValAddr":"axelarvaloper","bech32PrefixValPub":"axelarvaloperpub","bech32PrefixConsAddr":"axelarvalcons","bech32PrefixConsPub":"axelarvalconspub"},"bip44":{"coinType":118},"currencies":[{"coinDenom":"AXL","coinMinimalDenom":"uaxl","coinDecimals":6}],"feeCurrencies":[{"coinDenom":"AXL","coinMinimalDenom":"uaxl","coinDecimals":6}],"gasPriceStep":{"low":0.05,"average":0.125,"high":0.2},"features":["stargate","no-legacy-stdTx","ibc-transfer"]}'

export default ({}) => {

    const [value, setValue] = useState(placeholder);

    return <>
        <label 
            htmlFor="keplr_wallet_json_configuration" 
            className="block mb-2 text-sm font-medium text-gray-900 dark:text-gray-400"
        >
            Keplr Wallet stringified JSON configuration (Axelar testnet placeholder below)
        </label>
        <textarea 
            id="keplr_wallet_json_configuration" 
            rows="4"
            value={value}
            onChange={e => setValue(e.target.value)}
            className="block p-2.5 w-full text-sm text-gray-900 bg-gray-50 rounded-lg border border-gray-300 focus:ring-blue-500 focus:border-blue-500 dark:bg-gray-700 dark:border-gray-600 dark:placeholder-gray-400 dark:text-white dark:focus:ring-blue-500 dark:focus:border-blue-500" 
            placeholder={placeholder}>
        </textarea>
    </>
};