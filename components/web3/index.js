import Image from "next/image";
import { useState, useEffect } from "react";
import { useSelector, useDispatch, shallowEqual } from "react-redux";
import Web3 from "web3";

import { equals_ignore_case } from "../../utils";
import evm_chains from "../../data/evm_chains.json";
import { CHAIN_ID } from "../../reducers/types";

export default ({
  environment = "mainnet",
  chain,
  symbol,
  image,
  address,
  decimals,
}) => {
  const _evm_chains = evm_chains?.[environment] || [];

  const dispatch = useDispatch();
  const { _chain_id } = useSelector(state => ({ _chain_id: state.chain_id }), shallowEqual);
  const { chain_id } = { ..._chain_id };

  const [web3, setWeb3] = useState(null);
  const [chainId, setChainId] = useState(null);
  const [data, setData] = useState(null);

  useEffect(() => {
    if (!web3) {
      setWeb3(new Web3(Web3.givenProvider));
    }
    else {
      try {
        web3.currentProvider._handleChainChanged = e => {
          try {
            const chainId = Web3.utils.hexToNumber(e?.chainId);
            setChainId(chainId);
            dispatch({
              type: CHAIN_ID,
              value: chainId,
            });
          } catch (error) {}
        };
      } catch (error) {}
    }
  }, [web3]);

  useEffect(() => {
    if (chain_id) {
      setChainId(chain_id);
    }
  }, [chain_id]);

  useEffect(() => {
    if (data?.chain_id === chainId && data?.contract) {
      addToken(data.chain_id, data.contract);
    }
  }, [chainId, data]);

  const addToken = async (chain_id, contract) => {
    if (web3 && contract) {
      if (chain_id === chainId) {
        try {
          const response = await web3.currentProvider.request({
            method: "wallet_watchAsset",
            params: {
              type: "ERC20",
              options: {
                address: contract.address,
                symbol: contract.symbol,
                decimals: contract.decimals,
                image: contract.image ? `${window.location.origin}${contract.image}` : undefined,
              },
            },
          });
        } catch (error) {}
        setData(null);
      }
      else {
        switchChain(chain_id, contract);
      }
    }
  };

  const switchChain = async (chain_id, contract) => {
    try {
      await web3.currentProvider.request({
        method: "wallet_switchEthereumChain",
        params: [{ chainId: web3.utils.toHex(chain_id) }],
      });
    } catch (error) {
      if (error.code === 4902) {
        try {
          await web3.currentProvider.request({
            method: "wallet_addEthereumChain",
            params: _evm_chains.find(c => c.chain_id === chain_id)?.provider_params,
          });
        } catch (error) {}
      }
    }
    if (contract) {
      setData({ chain_id, contract });
    }
  };

  return (
    <button
      onClick={() => {
        if (chain) {
          const chain_data = _evm_chains.find(c => equals_ignore_case(c.id, chain));
          if (symbol && address && decimals) {
            addToken(
              chain_data?.chain_id,
              {
                symbol,
                address,
                decimals,
                image: image || `/images/assets/${(symbol.startsWith("axl") && !symbol.endsWith("axl") ? symbol.replace("axl", "") : symbol).toLowerCase()}.png`,
              }
            );
          }
          else {
            switchChain(chain_data?.chain_id);
          }
        }
      }}
      className="min-w-max bg-gray-100 hover:bg-gray-200 dark:bg-gray-900 dark:hover:bg-gray-800 rounded-lg cursor-pointer flex items-center py-1.5 px-2"
    >
      <Image
        src="/images/wallets/metamask.png"
        alt=""
        width={16}
        height={16}
      />
    </button>
  );
};