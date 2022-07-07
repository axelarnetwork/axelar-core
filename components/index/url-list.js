import Image from "next/image";
import Link from "next/link";
import { BsFileEarmarkTextFill, BsStack } from "react-icons/bs";

const items = [
  {
    title: "About Axelar",
    icon: (
      <>
        <div className="flex dark:hidden items-center">
          <Image
            src="/logo/logo.png"
            alt=""
            width={24}
            height={24}
          />
        </div>
        <div className="hidden dark:flex items-center">
          <Image
            src="/logo/logo_dark.png"
            alt=""
            width={24}
            height={24}
          />
        </div>
      </>
    ),
    url: "/learn",
    external: false,
  },
  {
    title: "Whitepaper",
    icon: (
      <BsFileEarmarkTextFill size={24} />
    ),
    url: "https://axelar.network/wp-content/uploads/2021/07/axelar_whitepaper.pdf",
    external: true,
  },
  {
    title: "Resources",
    icon: (
      <BsStack size={24} />
    ),
    url: "/resources",
    external: false,
  },
];

export default () => {
  return (
    <div className="grid grid-flow-row grid-cols-1 sm:grid-cols-3 gap-4 sm:gap-8 mt-6">
      {items.map((item, i) => {
        const {
          icon,
          title,
          url,
          external,
        } = { ...item };
        const element = (
          <div className="card-index">
            <div className="flex items-center space-x-3">
              {icon}
              <span className="text-base font-semibold">
                {title}
              </span>
            </div>
          </div>
        );

        return external ?
          <a
            key={i}
            href={url}
            target="_blank"
            rel="noopenner noreferrer"
            className="no-underline text-black dark:text-white"
          >
            {element}
          </a>
          :
          <Link
            key={i}
            href={url}
          >
            <a className="no-underline text-black dark:text-white">
              {element}
            </a>
          </Link>
      })}
    </div>
  );
};