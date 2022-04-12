import Image from "next/image";
import { useState, useEffect } from "react";
import { useSelector, useDispatch, shallowEqual } from "react-redux";
import Web3 from "web3";

import { CHAIN_ID_DATA } from "../../reducers/types";
import chains_data from "../../data/evm_chains.json";

export default ({ environment = "mainnet", chain, symbol, image, address, decimals }) => {
  const dispatch = useDispatch();
  const { chain_id } = useSelector(state => ({ chain_id: state.chain_id }), shallowEqual);
  const { chain_id_data } = { ...chain_id };

  const [web3, setWeb3] = useState(null);
  const [chainId, setChainId] = useState(null);
  const [addTokenData, setAddTokenData] = useState(null);

  useEffect(() => {
    if (!web3) {
      setWeb3(new Web3(Web3.givenProvider));
    }
    else {
      try {
        web3.currentProvider._handleChainChanged = e => {
          try {
            setChainId(Web3.utils.hexToNumber(e?.chainId));
            dispatch({
              type: CHAIN_ID_DATA,
              value: Web3.utils.hexToNumber(e?.chainId),
            });
          } catch (error) {}
        }
      } catch (error) {}
    }
  }, [web3]);

  useEffect(() => {
    if (chain_id_data) {
      setChainId(chain_id_data);
    }
  }, [chain_id_data]);

  useEffect(() => {
    if (addTokenData?.chain_id === chainId && addTokenData?.contract) {
      addTokenToMetaMask(addTokenData.chain_id, addTokenData.contract);
    }
  }, [chainId, addTokenData]);

  const addTokenToMetaMask = async (chain_id, contract) => {
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
                decimals: Number(contract.decimals),
                image: contract.image ? `${window.location.origin}${contract.image}` : undefined,
              },
            },
          });
        } catch (error) {}
        setAddTokenData(null);
      }
      else {
        switchNetwork(chain_id, contract);
      }
    }
  };

  const switchNetwork = async (chain_id, contract) => {
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
            params: chains_data?.[environment]?.find(c => c.chain_id === chain_id)?.provider_params,
          });
        } catch (error) {}
      }
    }

    if (contract) {
      setAddTokenData({ chain_id, contract });
    }
  };

  return (
    <button
      onClick={() => {
        if (chain) {
          const chain_data = chains_data?.[environment]?.find(c => c.id?.toLowerCase() === chain.toLowerCase());
          if (symbol && address && decimals) {
            addTokenToMetaMask(
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
            switchNetwork(chain_data?.chain_id);
          }
        }
      }}
      className="bg-gray-100 hover:bg-gray-200 dark:bg-gray-900 dark:hover:bg-gray-800 rounded-lg cursor-pointer flex items-center py-1.5 px-2"
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