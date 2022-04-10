import { useState, useEffect } from "react";
import Image from "next/image";

import Web3 from "web3";

const chains_data = {
  mainnet: [
    {
      id: "ethereum",
      name: "Ethereum",
      chain_id: 1,
      image: "/images/chains/ethereum.png",
      provider_params: [
        {
          chainId: "0x1",
          chainName: "Ethereum Mainnet",
          rpcUrls: ["https://rpc.ankr.com/eth"],
          nativeCurrency: {
            name: "Ether",
            symbol: "ETH",
            decimals: 18,
          },
          blockExplorerUrls: ["https://etherscan.io"],
        },
      ],
    },
    {
      id: "avalanche",
      name: "Avalache",
      chain_id: 43114,
      image: "/images/chains/avalache.png",
      provider_params: [
        {
          chainId: "0xa86a",
          chainName: "Avalanche Mainnet C-Chain",
          rpcUrls: ["https://api.avax.network/ext/bc/C/rpc"],
          nativeCurrency: {
            name: "Avalanche",
            symbol: "AVAX",
            decimals: 18,
          },
          blockExplorerUrls: ["https://snowtrace.io"],
        },
      ],
    },
    {
      id: "polygon",
      name: "Polygon",
      chain_id: 137,
      image: "/images/chains/polygon.png",
      provider_params: [
        {
          chainId: "0x89",
          chainName: "Matic Mainnet",
          rpcUrls: ["https://polygon-rpc.com", "https://matic-mainnet.chainstacklabs.com", "https://rpc-mainnet.maticvigil.com"],
          nativeCurrency: {
            name: "Matic",
            symbol: "MATIC",
            decimals: 18,
          },
          blockExplorerUrls: ["https://polygonscan.com"],
        },
      ],
    },
    {
      id: "fantom",
      name: "Fantom",
      chain_id: 250,
      image: "/images/chains/fantom.png",
      provider_params: [
        {
          chainId: "0xfa",
          chainName: "Fantom Opera",
          rpcUrls: ["https://rpc.ftm.tools", "https://rpc.ankr.com/fantom", "https://rpcapi.fantom.network"],
          nativeCurrency: {
            name: "Fantom",
            symbol: "FTM",
            decimals: 18,
          },
          blockExplorerUrls: ["https://ftmscan.com"],
        },
      ],
    },
    {
      id: "moonbeam",
      name: "Moonbeam",
      chain_id: 1284,
      image: "/images/chains/moonbeam.png",
      provider_params: [
        {
          chainId: "0x504",
          chainName: "Moonbeam",
          rpcUrls: ["https://rpc.api.moonbeam.network"],
          nativeCurrency: {
            name: "Glimmer",
            symbol: "GLMR",
            decimals: 18,
          },
          blockExplorerUrls: ["https://moonscan.io"],
        },
      ],
    },
  ],
  testnet: [
    {
      id: "ethereum",
      name: "Ethereum Ropsten",
      chain_id: 3,
      image: "/images/chains/ethereum.png",
      provider_params: [
        {
          chainId: "0x3",
          chainName: "Ethereum Ropsten",
          rpcUrls: ["https://ropsten.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"],
          nativeCurrency: {
            name: "Ropsten Ether",
            symbol: "ROP",
            decimals: 18,
          },
          blockExplorerUrls: ["https://ropsten.etherscan.io"],
        },
      ],
    },
    {
      id: "avalanche",
      name: "Avalache Fuji",
      chain_id: 43113,
      image: "/images/chains/avalache.png",
      provider_params: [
        {
          chainId: "0xa869",
          chainName: "Avalanche Testnet C-Chain",
          rpcUrls: ["https://api.avax-test.network/ext/bc/C/rpc"],
          nativeCurrency: {
            name: "Avalanche",
            symbol: "AVAX",
            decimals: 18,
          },
          blockExplorerUrls: ["https://testnet.snowtrace.io"],
        },
      ],
    },
    {
      id: "polygon",
      name: "Polygon Mumbai",
      chain_id: 80001,
      image: "/images/chains/polygon.png",
      provider_params: [
        {
          chainId: "0x13881",
          chainName: "Polygon Mumbai",
          rpcUrls: ["https://rpc-mumbai.maticvigil.com", "https://rpc-mumbai.matic.today", "https://matic-mumbai.chainstacklabs.com"],
          nativeCurrency: {
            name: "Matic",
            symbol: "MATIC",
            decimals: 18,
          },
          blockExplorerUrls: ["https://mumbai.polygonscan.com"],
        },
      ],
    },
    {
      id: "fantom",
      name: "Fantom Testnet",
      chain_id: 4002,
      image: "/images/chains/fantom.png",
      provider_params: [
        {
          chainId: "0xfa2",
          chainName: "Fantom Testnet",
          rpcUrls: ["https://rpc.testnet.fantom.network"],
          nativeCurrency: {
            name: "Fantom",
            symbol: "FTM",
            decimals: 18,
          },
          blockExplorerUrls: ["https://testnet.ftmscan.com"],
        },
      ],
    },
    {
      id: "moonbeam",
      name: "Moonbase Alpha",
      chain_id: 1287,
      image: "/images/chains/moonbeam.png",
      provider_params: [
        {
          chainId: "0x507",
          chainName: "Moonbase Alpha",
          rpcUrls: ["https://rpc.api.moonbase.moonbeam.network"],
          nativeCurrency: {
            name: "Dev",
            symbol: "DEV",
            decimals: 18,
          },
          blockExplorerUrls: ["https://moonbase.moonscan.io"],
        },
      ],
    },
  ],
};

export default ({ environment = "mainnet", chain, symbol, address, decimals }) => {
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
          } catch (error) {}
        }
      } catch (error) {}
    }
  }, [web3])

  useEffect(() => {
    if (addTokenData?.chain_id === chainId && addTokenData?.contract) {
      addTokenToMetaMask(addTokenData.chain_id, addTokenData.contract);
    }
  }, [chainId, addTokenData])

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
                symbol: symbol,
                address,
                decimals,
                image: `/images/assets/${(symbol.startsWith('axl') && !symbol.endsWith('axl') ? symbol.replace('axl', '') : symbol).toLowerCase()}.png`,
              }
            );
          }
          else {
            switchNetwork(chain_data?.chain_id);
          }
        }
      }}
      className="bg-gray-200 dark:bg-gray-800 rounded-lg cursor-pointer flex items-center py-1.5 px-2"
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