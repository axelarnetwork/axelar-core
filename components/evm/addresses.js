import Image from "next/image";
import { useState } from "react";

import Dropdown from "../dropdown";
import AddToWeb3 from "../web3";
import Copy from "../copy";
import utils from "../../utils";
import evm_chains from "../../data/evm_chains.json";
import evm_assets from "../../data/evm_assets.json";
import gateways from "../../data/gateways.json";
import ibc_assets from "../../data/ibc_assets.json";

export default ({ environment = "mainnet" }) => {
  const [chainData, setChainData] = useState(null);
  const [assetData, setAssetData] = useState(null);

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-center space-x-3">
        <Dropdown
          environment={environment}
          dataName="evm_chains"
          placeholder="Select Chain"
          hasAllOptions={true}
          allOptionsName="All Chains"
          defaultSelectedKey={chainData?.id || ""}
          onSelect={c => {
            setChainData(c);
            if (c === "" && !assetData) {
              setAssetData(null);
            }
          }}
        />
        <Dropdown
          environment={environment}
          chain={chainData?.id}
          dataName="evm_assets"
          placeholder="Select Asset"
          hasAllOptions={true}
          allOptionsName="All Assets"
          defaultSelectedKey={assetData?.id || (chainData && "") || assetData}
          onSelect={a => setAssetData(a)}
        />
      </div>
      <div className="grid grid-flow-row grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
        {(chainData || !assetData) && evm_chains?.[environment]?.filter(c => !chainData || c?.id === chainData.id).map((c, key) => {
          const rpc_url = c.provider_params?.[0]?.rpcUrls?.[0];
          const explorer_url = c.provider_params?.[0]?.blockExplorerUrls?.[0];
          const gateway_contract_address = gateways?.[environment]?.find(g => g?.id === c.id)?.address;

          return (
            <div key={key} className="card h-full flex flex-col justify-between p-4">
              <div className="flex items-start justify-between space-x-2">
                <div className="flex items-start space-x-2">
                  <Image
                    src={c.image}
                    alt=""
                    width={24}
                    height={24}
                    className="rounded-full"
                  />
                  <div className="flex flex-col">
                    <span className="text-sm font-semibold">{c.name}</span>
                    <span className="text-gray-400 dark:text-gray-500 text-xs font-medium">Chain ID: {c.chain_id}</span>
                  </div>
                </div>
                <AddToWeb3 environment={environment} chain={c.id} />
              </div>
              <div className="flex flex-col text-xs space-y-3 mt-3">
                <div className="flex flex-col">
                  <div className="flex items-center space-x-1">
                    <span className="text-gray-500 dark:text-gray-300">RPC URL</span>
                    {rpc_url && (
                      <Copy value={rpc_url} />
                    )}
                  </div>
                  {rpc_url ?
                    <a
                      href={rpc_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="no-underline text-blue-500 dark:text-white font-semibold"
                    >
                      {utils.ellipse(rpc_url, 18)}
                    </a>
                    :
                    <span className="text-gray-400 dark:text-gray-600">-</span>
                  }
                </div>
                <div className="flex flex-col">
                  <div className="flex items-center space-x-1">
                    <span className="text-gray-500 dark:text-gray-300">Block Explorer URL</span>
                    {explorer_url && (
                      <Copy value={explorer_url} />
                    )}
                  </div>
                  {explorer_url ?
                    <a
                      href={explorer_url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="no-underline text-blue-500 dark:text-white font-semibold"
                    >
                      {utils.ellipse(explorer_url, 18)}
                    </a>
                    :
                    <span className="text-gray-400 dark:text-gray-600">-</span>
                  }
                </div>
                <div className="flex flex-col">
                  <div className="flex items-center space-x-1">
                    <span className="text-gray-500 dark:text-gray-300">Gateway Contract</span>
                    {gateway_contract_address && (
                      <Copy value={gateway_contract_address} />
                    )}
                  </div>
                  {gateway_contract_address ?
                    <a
                      href={`${explorer_url}/address/${gateway_contract_address}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="no-underline text-blue-500 dark:text-blue-300 font-semibold"
                    >
                      {utils.ellipse(gateway_contract_address, 14)}
                    </a>
                    :
                    <span className="text-gray-400 dark:text-gray-600">-</span>
                  }
                </div>
              </div>
            </div>
          );
        })}
        {(chainData || assetData || assetData === "") && evm_assets?.[environment]?.filter(a => !assetData || a?.id === assetData.id).flatMap(a => a?.contracts?.map(c => {
          return {
            ...a,
            ...c,
          };
        }).filter(a => !chainData || a.chain?.toLowerCase() === chainData.id?.toLowerCase()) || []).map((a, key) => {
          const chain_data = evm_chains?.[environment]?.find(c => c?.id === a.chain);
          const explorer_url = chain_data?.provider_params?.[0]?.blockExplorerUrls?.[0];
          const token_contract_address = a.address;
          const denom = a.id;
          const ethereum_transfer_fee = evm_assets[environment].find(_a => _a?.id === a.id)?.contracts?.find(c => c?.chain === "ethereum")?.transfer_fee;
          const non_ethereum_transfer_fee = evm_assets[environment].find(_a => _a?.id === a.id)?.transfer_fee;
          const cosmos_transfer_fee = ibc_assets[environment].find(_a => _a?.id === a.id)?.transfer_fee;

          return (
            <div key={key} className="card h-full flex flex-col justify-between p-4">
              <div className="flex items-start justify-between space-x-2">
                <div className="flex items-start space-x-2">
                  <Image
                    src={a.image}
                    alt=""
                    width={24}
                    height={24}
                    className="rounded-full"
                  />
                  <div className="flex flex-col">
                    <div className="flex items-center space-x-2">
                      <div className="flex items-center space-x-1.5">
                        <span className="whitespace-nowrap text-xs font-semibold">{a.name}</span>
                        <span className="text-gray-400 dark:text-gray-500 text-xs font-medium">{a.symbol}</span>
                      </div>
                      {!chainData && chain_data && (
                        <div className="min-w-max flex items-center">
                          <Image
                            src={chain_data.image}
                            alt=""
                            width={16}
                            height={16}
                            className="rounded-full"
                          />
                        </div>
                      )}
                    </div>
                    <span className="whitespace-nowrap text-gray-400 dark:text-gray-500 text-xs font-medium">Decimals: {a.decimals}</span>
                  </div>
                </div>
                <AddToWeb3 environment={environment} { ...a } />
              </div>
              <div className="flex flex-col text-xs space-y-3 mt-3">
                <div className="flex flex-col">
                  <div className="flex items-center space-x-1">
                    <span className="text-gray-500 dark:text-gray-300">Token Contract</span>
                    {token_contract_address && (
                      <Copy value={token_contract_address} />
                    )}
                  </div>
                  {token_contract_address ?
                    <a
                      href={`${explorer_url}/address/${token_contract_address}`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="no-underline text-blue-500 dark:text-blue-300 font-semibold"
                    >
                      {utils.ellipse(token_contract_address, 14)}
                    </a>
                    :
                    <span className="text-gray-400 dark:text-gray-600">-</span>
                  }
                </div>
                <div className="flex flex-col space-y-1">
                  <span className="text-gray-500 dark:text-gray-300">Transfer Fee</span>
                  <div className="grid grid-flow-row grid-cols-2 gap-2">
                    <div className="flex flex-col">
                      <div className="flex items-center space-x-1">
                        <span className="text-gray-500 dark:text-gray-300">to Ethereum</span>
                      </div>
                      {ethereum_transfer_fee ?
                        <span className="font-semibold">
                          {ethereum_transfer_fee} {a.symbol}
                        </span>
                        :
                        <span className="text-gray-400 dark:text-gray-600">-</span>
                      }
                    </div>
                    <div className="flex flex-col">
                      <div className="flex items-center space-x-1">
                        <span className="text-gray-500 dark:text-gray-300">to non-Ethereum</span>
                      </div>
                      {non_ethereum_transfer_fee ?
                        <span className="font-semibold">
                          {non_ethereum_transfer_fee} {a.symbol}
                        </span>
                        :
                        <span className="text-gray-400 dark:text-gray-600">-</span>
                      }
                    </div>
                  </div>
                  <div className="grid grid-flow-row grid-cols-2 gap-2">
                    <div className="flex flex-col">
                      <div className="flex items-center space-x-1">
                        <span className="text-gray-500 dark:text-gray-300">to Cosmos</span>
                      </div>
                      {cosmos_transfer_fee ?
                        <span className="font-semibold">
                          {cosmos_transfer_fee} {a.symbol}
                        </span>
                        :
                        <span className="text-gray-400 dark:text-gray-600">-</span>
                      }
                    </div>
                    {/*<div className="flex flex-col">
                      <div className="flex items-center space-x-1">
                        <span className="text-gray-500 dark:text-gray-300">Denom</span>
                        {denom && (
                          <Copy value={denom} />
                        )}
                      </div>
                      {denom ?
                        <span className="font-semibold">
                          {denom}
                        </span>
                        :
                        <span className="text-gray-400 dark:text-gray-600">-</span>
                      }
                    </div>*/}
                  </div>
                </div>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
};