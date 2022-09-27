import Image from "next/image";
import _ from "lodash";
import { BiTransfer } from "react-icons/bi";

import cosmos_chains from "../../data/cosmos_chains.json";
import ibc_channels from "../../data/ibc_channels.json";

const MAIN_CHAIN = "axelarnet";

export default ({ environment = "mainnet" }) => {
  const _cosmos_chains = cosmos_chains?.[environment] || [];
  const _ibc_channels = ibc_channels?.[environment] || [];

  const pairs =
    Object.entries(
      _.groupBy(
        _ibc_channels.map(c => {
          return {
            ...c,
            other_chain: _.head([c?.from, c?.to].filter(cid => cid && cid !== MAIN_CHAIN)),
          };
        }),
        "other_chain",
      )
    )
    .map(([other_chain, channels]) => {
      return {
        chain_data: _cosmos_chains.find(c => c?.id === MAIN_CHAIN),
        other_chain_data: _cosmos_chains.find(c => c?.id === other_chain),
        channels,
      };
    });

  return (
    <div className="grid grid-flow-row grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
      {pairs.map((p, i) => {
        const {
          chain_data,
          other_chain_data,
          channels,
        } = { ...p };

        return (
          <div
            key={i}
            className="border dark:border-gray-700 rounded-xl grid grid-cols-3 gap-4 p-3"
          >
            <div className="flex flex-col items-center space-y-1">
              {chain_data?.image && (
                <Image
                  src={chain_data.image}
                  alt=""
                  width={32}
                  height={32}
                  className="rounded-full"
                />
              )}
              <span className="text-xs font-semibold">
                {chain_data?.name}
              </span>
            </div>
            <div className="flex flex-col items-center">
              <span className="whitespace-nowrap text-gray-600 dark:text-gray-400 text-xs">
                {channels.find(c => c?.from === chain_data?.id)?.channel_id}
              </span>
              <BiTransfer size={20} />
              <span className="whitespace-nowrap text-gray-600 dark:text-gray-400 text-xs">
                {channels.find(c => c?.from === other_chain_data?.id)?.channel_id}
              </span>
            </div>
            <div className="flex flex-col items-center space-y-1">
              {other_chain_data?.image && (
                <Image
                  src={other_chain_data.image}
                  alt=""
                  width={32}
                  height={32}
                  className={`${['fetch'].includes(other_chain_data.id) ?'bg-black rounded-full' : ''}`}
                />
              )}
              <span className="text-xs font-semibold">
                {other_chain_data?.name}
              </span>
            </div>
          </div>
        );
      })}
    </div>
  );
};