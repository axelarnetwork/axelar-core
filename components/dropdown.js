import Image from "next/image";
import { Fragment, useState, useEffect } from "react";
import _ from "lodash";
import { Menu, Transition } from "@headlessui/react";
import { BiChevronDown, BiChevronUp } from "react-icons/bi";

import evm_chains from "../data/evm_chains.json";
import evm_assets from "../data/evm_assets.json";
import gateways from "../data/gateways.json";
import cosmos_chains from "../data/cosmos_chains.json";
import ibc_assets from "../data/ibc_assets.json";

const data = {
  evm_chains,
  evm_assets,
  gateways,
  cosmos_chains,
  ibc_assets,
};

export default ({ environment, chain, dataName, placeholder, hasAllOptions, allOptionsName = "All", defaultSelectedKey, onSelect, align = "left", className = "" }) => {
  const [options, setOptions] = useState(null);
  const [selectedKey, setSelectedKey] = useState(null);

  useEffect(() => {
    let _options;
    switch (dataName) {
      case "evm_chains":
        _options = data[dataName]?.[environment];
        break;
      case "evm_assets":
        _options = data[dataName]?.[environment]?.flatMap(o => {
          const contracts = o?.contracts?.filter(c => !chain || c?.chain?.toLowerCase() === chain.toLowerCase()).filter((c, i) => chain || i < 1) || [];
          return contracts.map(c => {
            return {
              ...o,
              ...c,
            };
          });
        }).map(o => {
          return {
            ...o,
            name: o?.symbol,
          };
        });
        break;
      case "chains":
        _options = _.concat(data.evm_chains?.[environment] || [], data.cosmos_chains?.[environment] || []);
        break;
      case "assets":
        _options = _.uniqBy(_.concat(data.evm_assets?.[environment] || [], data.ibc_assets?.[environment] || []), 'id');
        break;
      default:
        _options = data[dataName];
        break;
    };

    setOptions(_options || []);
  }, [environment, chain, dataName]);

  useEffect(() => {
    setSelectedKey(defaultSelectedKey);
  }, [defaultSelectedKey]);

  const selectedData = options?.find(o => o?.id === selectedKey) || selectedKey;

  return (
    <Menu as="div" className={`relative inline-block text-left ${className}`}>
      {({ open }) => (
        <>
          <div>
            <Menu.Button
              onClick={() => setOpen(!open)}
              className="bg-white dark:bg-dark hover:bg-gray-50 dark:hover:bg-gray-900 w-full rounded-md border border-gray-300 dark:border-gray-700 shadow-sm focus:outline-none inline-flex justify-center text-sm font-medium text-gray-900 dark:text-gray-100 py-2 px-4"
            >
              {selectedData ?
                <div className="flex items-center space-x-2">
                  {selectedData.image && (
                    <Image
                      src={selectedData.image}
                      alt=""
                      width={24}
                      height={24}
                      className="rounded-full"
                    />
                  )}
                  <span className="font-bold">{selectedData.name}</span>
                </div>
                :
                selectedData === "" ?
                  <span className="font-bold">{allOptionsName}</span>
                  :
                  placeholder || "Select Options"
              }
              {open ?
                <BiChevronUp size={selectedData?.image ? 24 : 20} className="text-gray-800 dark:text-gray-200 ml-1.5 -mr-1.5" />
                :
                <BiChevronDown size={selectedData?.image ? 24 : 20} className="text-gray-800 dark:text-gray-200 ml-1.5 -mr-1.5" />
              }
            </Menu.Button>
          </div>
          <Transition
            as={Fragment}
            enter="transition ease-out duration-100"
            enterFrom="transform opacity-0 scale-95"
            enterTo="transform opacity-100 scale-100"
            leave="transition ease-in duration-75"
            leaveFrom="transform opacity-100 scale-100"
            leaveTo="transform opacity-0 scale-95"
          >
            <Menu.Items className={`bg-white w-48 min-w-max dark:bg-dark absolute z-10 rounded-md shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none origin-top-${align} ${align}-0 mt-2`}>
              <div className="py-1">
                {hasAllOptions && (
                  <Menu.Item key={-1}>
                    {({ active }) => (
                      <div
                        onClick={() => {
                          setSelectedKey("");
                          if (onSelect) {
                            onSelect("");
                          }
                        }}
                        className={`${active ? "bg-gray-100 dark:bg-gray-900 text-dark dark:text-white" : "text-gray-800 dark:text-gray-200"} ${selectedKey === "" ? "font-bold" : active ? "font-semibold" : "font-medium"} cursor-pointer flex items-center text-sm space-x-2 py-2 px-4`}
                      >
                        <span>{allOptionsName}</span>
                      </div>
                    )}
                  </Menu.Item>
                )}
                {options?.map((option, key) => (
                  <Menu.Item key={key}>
                    {({ active }) => (
                      <div
                        onClick={() => {
                          setSelectedKey(option.id);
                          if (onSelect) {
                            onSelect(options?.find(o => o?.id === option.id));
                          }
                        }}
                        className={`${active ? "bg-gray-100 dark:bg-gray-900 text-dark dark:text-white" : "text-gray-800 dark:text-gray-200"} ${selectedKey === option.id ? "font-bold" : active ? "font-semibold" : "font-medium"} cursor-pointer flex items-center text-sm space-x-2 py-2 px-4`}
                      >
                        {option.image && (
                          <Image
                            src={option.image}
                            alt=""
                            width={24}
                            height={24}
                            className="rounded-full"
                          />
                        )}
                        <span>{option.name}</span>
                      </div>
                    )}
                  </Menu.Item>
                ))}
              </div>
            </Menu.Items>
          </Transition>
        </>
      )}
    </Menu>
  )
};