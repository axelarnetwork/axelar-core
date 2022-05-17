import Image from "next/image";
import { useState } from "react";

import Dropdown from "../dropdown";
import Copy from "../copy";
import evm_chains from "../../data/evm_chains.json";
import evm_assets from "../../data/evm_assets.json";
import cosmos_chains from "../../data/cosmos_chains.json";
import ibc_assets from "../../data/ibc_assets.json";

export default ({ environment = "mainnet" }) => {
  const [assetData, setAssetData] = useState(evm_assets?.[environment]?.find(a => a?.id === "uusdc"));
  const [sourceChainData, setSourceChainData] = useState(evm_chains?.[environment]?.find(c => c?.id === "avalanche"));
  const [destinationChainData, setDestinationChainData] = useState(cosmos_chains?.[environment]?.find(c => c?.id === "osmosis"));

  const getTransferFee = chain => {
    let transfer_fee;
    if (chain && assetData?.id) {
      if (evm_chains?.[environment]?.findIndex(c => c?.id === chain) > -1) {
        const chain_data = evm_chains[environment].find(c => c?.id === chain);
        const asset_data = evm_assets?.[environment]?.find(a => a?.id === assetData.id);
        transfer_fee = asset_data?.contracts?.find(c => c?.chain === chain)?.transfer_fee || (asset_data?.contracts?.findIndex(c => c?.chain === chain) > -1 ? asset_data?.transfer_fee : null);
      }
      else if (cosmos_chains?.[environment]?.findIndex(c => c?.id === chain) > -1) {
        const asset_data = ibc_assets?.[environment]?.find(a => a?.id === assetData.id);
        transfer_fee = asset_data?.transfer_fee;
      }
    }
    return transfer_fee;
  };

  const { symbol } = { ...assetData };
  const sourceTransferFee = getTransferFee(sourceChainData?.id);
  const destinationTransferFee = getTransferFee(destinationChainData?.id);
  const totalFee = parseFloat(((sourceTransferFee || 0) + (destinationTransferFee || 0)).toFixed(6));
  return (
    <div className="max-w-lg border dark:border-gray-500 rounded-2xl shadow dark:shadow-gray-500 flex flex-col p-6">
      <div className="grid grid-cols-3 items-center gap-6">
        <span className="text-base font-bold">
          Asset
        </span>
        <Dropdown
          environment={environment}
          dataName="assets"
          placeholder="Select Asset"
          defaultSelectedKey={assetData?.id || ""}
          onSelect={a => setAssetData(a)}
          className="min-w-max"
        />
        <span className="whitespace-nowrap text-base font-semibold text-right">
          Fee
        </span>
      </div>
      <div className="grid grid-cols-3 items-center gap-6 mt-4">
        <span className="text-base font-bold">
          Source chain
        </span>
        <Dropdown
          environment={environment}
          dataName="chains"
          placeholder="Select Chain"
          defaultSelectedKey={sourceChainData?.id || ""}
          onSelect={c => setSourceChainData(c)}
          className="min-w-max"
        />
        <span className="whitespace-nowrap text-base font-semibold text-right">
          {sourceTransferFee || 'N/A'} {symbol}
        </span>
      </div>
      <div className="flex items-center justify-end">
        <span className="text-xl font-bold">
          +
        </span>
      </div>
      <div className="grid grid-cols-3 items-center gap-6 mt-1">
        <span className="text-base font-bold">
          Destination chain
        </span>
        <Dropdown
          environment={environment}
          dataName="chains"
          placeholder="Select Chain"
          defaultSelectedKey={destinationChainData?.id || ""}
          onSelect={c => setDestinationChainData(c)}
          className="min-w-max"
        />
        <span className="whitespace-nowrap text-base font-semibold text-right">
          {destinationTransferFee || 'N/A'} {symbol}
        </span>
      </div>
      <div className="border-t-2 dark:border-gray-500 mt-4" />
      <div className="flex items-center justify-between mt-3">
        <span className="text-lg font-bold">
          Total
        </span>
        <span className="text-xl font-bold">
          {totalFee} {symbol}
        </span>
      </div>
      <div className="h-3 border-y-2 dark:border-gray-500 mt-1" />
    </div>
  );
};