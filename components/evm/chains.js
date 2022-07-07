import Image from "next/image";

import AddChain from "../web3";
import Copy from "../copy";
import { ellipse } from "../../utils";
import evm_chains from "../../data/evm_chains.json";
import gateways from "../../data/gateways.json";
import gas_services from "../../data/gas_services.json";

export default ({ environment = "mainnet" }) => {
  const _evm_chains = evm_chains?.[environment] || [];
  const _gateways = gateways?.[environment] || [];
  const _gas_services = gas_services?.[environment] || [];

  return (
    <div className="grid grid-flow-row grid-cols-1 lg:grid-cols-2 gap-4">
      {_evm_chains.filter(c => !c?.is_staging).map((c, i) => {
        const {
          id,
          chain_id,
          name,
          provider_params,
          image,
        } = { ...c };
        const explorer_url = provider_params?.[0]?.blockExplorerUrls?.[0];
        const gateway_contract_address = _gateways.find(_c => _c?.id === id)?.address;
        const gas_service_address = _gas_services.find(_c => _c?.id === id)?.address;

        return (
          <div
            key={i}
            className="border dark:border-gray-700 rounded-xl flex flex-col justify-between space-y-2 p-4"
          >
            <div className="flex items-start justify-between space-x-2">
              <div className="flex items-center space-x-3">
                {image && (
                  <Image
                    src={image}
                    alt=""
                    width={32}
                    height={32}
                    className="rounded-full"
                  />
                )}
                <div className="flex flex-col">
                  <span className="text-base font-semibold">
                    {name}
                  </span>
                  <span className="text-gray-400 dark:text-gray-500 text-sm font-medium">
                    Chain ID: {c.chain_id}
                  </span>
                </div>
              </div>
              <AddChain
                environment={environment}
                chain={id}
              />
            </div>
            <div className="flex flex-col flex-wrap justify-between">
              <span className="whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                Gateway Contract:
              </span>
              <div className="flex items-center text-sm space-x-1">
                {gateway_contract_address ?
                  <a
                    href={`${explorer_url}/address/${gateway_contract_address}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="no-underline text-blue-500 dark:text-white font-semibold"
                  >
                    {ellipse(gateway_contract_address, 14)}
                  </a>
                  :
                  <span className="text-gray-500 dark:text-white font-semibold">
                    -
                  </span>
                }
                {gateway_contract_address && (
                  <Copy
                    value={gateway_contract_address}
                    size={18}
                  />
                )}
              </div>
            </div>
            <div className="flex flex-col flex-wrap justify-between">
              <span className="whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                Gas Service Contract:
              </span>
              <div className="flex items-center text-sm space-x-1">
                {gas_service_address ?
                  <a
                    href={`${explorer_url}/address/${gas_service_address}`}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="no-underline text-blue-500 dark:text-white font-semibold"
                  >
                    {ellipse(gas_service_address, 14)}
                  </a>
                  :
                  <span className="text-gray-500 dark:text-white font-semibold">
                    -
                  </span>
                }
                {gas_service_address && (
                  <Copy
                    value={gas_service_address}
                    size={18}
                  />
                )}
              </div>
            </div>
          </div>
        );
      })}
    </div>
  );
};