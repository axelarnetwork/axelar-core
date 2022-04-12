import Image from "next/image";
import _ from "lodash";
import { BiTransfer } from "react-icons/bi";

import cosmos_chains from "../../data/cosmos_chains.json";
import ibc_channels from "../../data/ibc_channels.json";

const main_chain = "axelarnet";

export default ({ environment = "mainnet" }) => {
  const pairs = Object.entries(_.groupBy(ibc_channels?.[environment]?.map(c => {
    return {
      ...c,
      other_chain: _.head([c?.from, c?.to].filter(cid => cid && cid !== main_chain)),
    };
  }) || [], "other_chain")).map(([other_chain, channels]) => {
    return {
      chain_data: cosmos_chains?.[environment]?.find(c => c?.id === main_chain),
      other_chain_data: cosmos_chains?.[environment]?.find(c => c?.id === other_chain),
      channels,
    };
  });

  return (
    <div className="grid grid-flow-row grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {pairs.map((p, key) => (
        <div key={key} className="shadow-lg rounded-xl flex items-center justify-between space-x-2 p-3">
          <div className="flex flex-col items-center space-y-1">
            <Image
              src={p.chain_data?.image}
              alt=""
              width={32}
              height={32}
              className="rounded-full"
            />
            <span className="font-semibold">{p.chain_data?.name}</span>
          </div>
          <div className="flex flex-col items-center">
            <span className="text-gray-600 dark:text-gray-400 text-sm">
              {p.channels.find(c => c?.from === p.chain_data?.id)?.channel_id}
            </span>
            <BiTransfer size={20} />
            <span className="text-gray-600 dark:text-gray-400 text-sm">
              {p.channels.find(c => c?.from === p.other_chain_data?.id)?.channel_id}
            </span>
          </div>
          <div className="flex flex-col items-center space-y-1">
            <Image
              src={p.other_chain_data?.image}
              alt=""
              width={32}
              height={32}
              className="rounded-full"
            />
            <span className="font-semibold">{p.other_chain_data?.name}</span>
          </div>
        </div>
      ))}
    </div>
  );
};