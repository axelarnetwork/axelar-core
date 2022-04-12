import { RiCodeFill, RiRadarFill, RiSettings4Fill, RiServerFill } from "react-icons/ri";
import { HiArrowNarrowRight } from "react-icons/hi";

const items = [
  {
    title: "Developer",
    description: "Use Axelar gateway contracts to call any EVM contract on any chain",
    icon: (
      <RiCodeFill size={24} />
    ),
    url: "/dev",
    external: false,
  },
  {
    title: "Satellite user",
    description: "Satellite is a web app built on top of the Axelar network. Use it to transfer assets from one chain to another.",
    icon: (
      <RiRadarFill size={24} />
    ),
    url: "/resources/satellite",
    external: false,
  },
  {
    title: "Node operator",
    description: "Learn how to run a node on the Axelar network",
    icon: (
      <RiSettings4Fill size={24} />
    ),
    url: "/node/join",
    external: false,
  },
  {
    title: "Validator",
    description: "Axelar validators facilitate cross-chain connections",
    icon: (
      <RiServerFill size={24} />
    ),
    url: "/validator/setup/overview",
    external: false,
  }
];

export default () => {
  return (
    <>
      <h2 className="border-0">Learn for your role</h2>
      <div className="grid grid-flow-row grid-cols-1 sm:grid-cols-3 gap-4 sm:gap-8 my-4">
        {items.map((item, key) => {
          const link = item.external ?
            <a
              href={item.url}
              target="_blank"
              rel="noopenner noreferrer"
              className="no-underline flex items-center space-x-1.5"
            >
              <span>Documentation</span>
              <HiArrowNarrowRight size={16} className="mt-0.5" />
            </a>
            :
            <a
              href={item.url}
              target="_blank"
              className="no-underline flex items-center space-x-1.5"
            >
              <span>Documentation</span>
              <HiArrowNarrowRight size={16} className="mt-0.5" />
            </a>

          const element = (
            <div className="card-index h-full flex flex-col justify-between">
              <div className="mb-2">
                <div className="flex items-center space-x-3">
                  {item.icon}
                  <span className="text-base font-semibold">{item.title}</span>
                </div>
                <div className="text-gray-500 dark:text-gray-400 mt-4">
                  {item.description}
                </div>
              </div>
              {link}
            </div>
          );

          return item.external ?
            <a
              key={key}
              href={item.url}
              target="_blank"
              rel="noopenner noreferrer"
              className="no-underline text-black dark:text-white"
            >
              {element}
            </a>
            :
            <a
              key={key}
              href={item.url}
              className="no-underline text-black dark:text-white"
            >
              {element}
            </a>
        })}
      </div>
    </>
  );
};