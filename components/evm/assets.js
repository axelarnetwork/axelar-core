import Image from "next/image";
import { useState } from "react";

import Dropdown from "../dropdown";
import AddToWeb3 from "../web3";
import Copy from "../copy";
import utils from "../../utils";
import evm_chains from "../../data/evm_chains.json";
import evm_assets from "../../data/evm_assets.json";

export default ({ environment = "mainnet" }) => {
  const [chainData, setChainData] = useState(null);
  const [assetData, setAssetData] = useState(evm_assets?.[environment]?.find(a => a?.id === "uusd"));

  const columns = [
    { id: "asset", title: "Asset" },
    { id: "chain", title: "Chain" },
    { id: "contract_address", title: "Contract address" },
    { id: "add_token", title: "", headerClassName: "text-right", className: "text-right" },
  ];

  const assets = evm_assets?.[environment]?.filter(a => !assetData || a?.id === assetData.id).flatMap(a => a?.contracts?.map(c => {
    return {
      ...a,
      ...c,
    };
  }).filter(a => !chainData || a.chain?.toLowerCase() === chainData.id?.toLowerCase()) || []) || [];

  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center justify-start space-x-3">
        <Dropdown
          environment={environment}
          dataName="evm_chains"
          placeholder="Select Chain"
          hasAllOptions={true}
          allOptionsName="All Chains"
          defaultSelectedKey={chainData?.id || ""}
          onSelect={c => {
            setChainData(c);
            if (c && evm_assets?.[environment]?.findIndex(a => (!assetData || a?.id === assetData.id) && a?.contracts?.findIndex(_c => _c?.chain === c?.id) > -1) < 0) {
              setAssetData("");
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
          defaultSelectedKey={assetData?.id || ""}
          onSelect={a => setAssetData(a)}
        />
      </div>
      <table className="max-w-fit block shadow rounded-lg overflow-x-auto">
        <thead className="bg-gray-100 dark:bg-black uppercase text-xs">
          <tr className="border-none">
            {columns.map((c, key) => (
              <th key={key} scope="col" className={`${key === 0 ? "rounded-tl-lg" : key === columns.length - 1 ? "rounded-tr-lg" : ""} border-none whitespace-nowrap font-bold py-3 px-4 ${c.headerClassName || ""}`}>
                {c.title}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {assets.map((a, key) => {
            const chain_data = evm_chains?.[environment]?.find(c => c?.id === a.chain);
            const explorer_url = chain_data?.provider_params?.[0]?.blockExplorerUrls?.[0];

            return (
              <tr key={key} className="border-none border-b">
                {columns.map((c, _key) => (
                  <th key={_key} scope="col" className={`${key % 2 === 0 ? "bg-transparent" : "bg-gray-50 dark:bg-black"} ${key === assets.length - 1 ? _key === 0 ? "rounded-bl-lg" : _key === columns.length - 1 ? "rounded-br-lg" : "" : ""} border-none whitespace-nowrap py-3 px-4 ${c.className || ""}`}>
                    {c.id === "asset" ?
                      <div className="min-w-max flex items-center space-x-3">
                        <Image
                          src={a.image}
                          alt=""
                          width={28}
                          height={28}
                          className="rounded-full"
                        />
                        <span className="whitespace-nowrap text-base font-semibold">{a.symbol}</span>
                      </div>
                      :
                      c.id === "chain" ?
                        <div className="min-w-max flex items-center space-x-2.5">
                          <Image
                            src={chain_data?.image}
                            alt=""
                            width={24}
                            height={24}
                            className="rounded-full"
                          />
                          <span className="whitespace-nowrap text-sm font-semibold">{chain_data?.name || a.chain}</span>
                        </div>
                        :
                        c.id === "contract_address" ?
                          <div className="flex items-center text-base space-x-1.5">
                            {a.address ?
                              <a
                                href={`${explorer_url}/address/${a.address}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="no-underline text-blue-500 dark:text-white font-medium"
                              >
                                {utils.ellipse(a.address, 24)}
                              </a>
                              :
                              <span className="text-gray-400 dark:text-gray-600">-</span>
                            }
                            {a.address && (
                              <Copy size={20} value={a.address} />
                            )}
                          </div>
                          :
                          c.id === "add_token" ?
                            <AddToWeb3 environment={environment} { ...a } />
                            :
                            null
                    }
                  </th>
                ))}
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
};