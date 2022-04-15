import Image from "next/image";

import AddToWeb3 from "../web3";
import Copy from "../copy";
import utils from "../../utils";
import evm_chains from "../../data/evm_chains.json";
import gateways from "../../data/gateways.json";

export default ({ environment = "mainnet" }) => {
  return (
    <div className="grid grid-flow-row grid-cols-1 lg:grid-cols-2 gap-4">
      {evm_chains?.[environment].map((c, key) => {
        const explorer_url = c.provider_params?.[0]?.blockExplorerUrls?.[0];
        const gateway_contract_address = gateways?.[environment]?.find(g => g?.id === c.id)?.address;

        return (
          <div key={key} className="border dark:border-gray-700 rounded-xl flex flex-col justify-between space-y-2 p-4">
            <div className="flex items-start justify-between space-x-2">
              <div className="flex items-center space-x-3">
                <Image
                  src={c.image}
                  alt=""
                  width={32}
                  height={32}
                  className="rounded-full"
                />
                <div className="flex flex-col">
                  <span className="text-base font-semibold">{c.name}</span>
                  <span className="text-gray-400 dark:text-gray-500 text-sm font-medium">Chain ID: {c.chain_id}</span>
                </div>
              </div>
              <AddToWeb3 environment={environment} chain={c.id} />
            </div>
            <div className="flex flex-wrap items-center justify-between">
              <span className="whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">Gateway Contract:</span>
              <div className="flex items-center text-sm space-x-1">
                {gateway_contract_address ?
                  <a
                    href={`${explorer_url}/address/${gateway_contract_address}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="no-underline text-blue-500 dark:text-white font-semibold"
                  >
                    {utils.ellipse(gateway_contract_address, 14)}
                  </a>
                  :
                  <span className="text-gray-400 dark:text-gray-600">-</span>
                }
                {gateway_contract_address && (
                  <Copy size={18} value={gateway_contract_address} />
                )}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};