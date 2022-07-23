import Link from "next/link";
import { AiFillBug, AiFillDatabase } from "react-icons/ai";
import { FiMonitor } from "react-icons/fi";
import { FaGasPump } from "react-icons/fa";
import { HiArrowNarrowRight } from "react-icons/hi";

const items = [
  {
    title: "5-min starter examples",
    description: "A curated list of representative examples in our axelar-local-gmp-examples repo",
    icon: (
      <AiFillDatabase size={24} />
    ),
    url: "/dev/build/5-min-starter-examples",
    external: false,
  },
  {
    title: "Gas & Executor Services",
    description: "Gas & Executor Services",
    icon: (
      <FaGasPump size={24} />
    ),
    url: "/dev/gas-services/intro",
    external: false,
  },
  {
    title: "Debug",
    description: "Debugging tools",
    icon: (
      <AiFillBug size={24} />
    ),
    url: "/dev/debug/error-debugging",
    external: false,
  },
  {
    title: "Monitor & recover",
    description: "Monitor & Recover",
    icon: (
      <FiMonitor size={24} />
    ),
    url: "/dev/monitor-recover/monitoring",
    external: false,
  }
];

export default () => {
  return (
    <div className="grid grid-flow-row grid-cols-1 gap-4 my-4 sm:grid-cols-2 sm:gap-8">
      {items.map((item, i) => {
        const {
          icon,
          title,
          description,
          url,
          external,
        } = { ...item };
        const link = (
          <div className="flex items-center text-blue-500 dark:text-blue-600 space-x-1.5">
            <span>
              Documentation
            </span>
            <HiArrowNarrowRight size={16} className="mt-0.5" />
          </div>
        );
        const element = (
          <div className="flex flex-col justify-between h-full card-index">
            <div className="mb-2">
              <div className="flex items-center space-x-3">
                {icon}
                <span className="text-base font-semibold">
                  {title}
                </span>
              </div>
              <div className="mt-4 text-gray-500 dark:text-gray-400">
                {description}
              </div>
            </div>
            {link}
          </div>
        );

        return (
          external ?
            <a
              key={i}
              href={url}
              target="_blank"
              rel="noopenner noreferrer"
              className="text-black no-underline dark:text-white"
            >
              {element}
            </a>
            :
            <Link
              key={i}
              href={url}
            >
              <a className="text-black no-underline dark:text-white">
                {element}
              </a>
            </Link>
        );
      })}
    </div>
  );
};