import Link from "next/link";
import Image from "next/image";

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
    url: "https://axelar.network",
    external: true,
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
    	{items.map((item, key) => {
    		const element = (
	    		<div className="card">
	    			<div className="flex items-center space-x-3">
	    				{item.icon}
	    				<span className="text-base font-semibold">{item.title}</span>
	    			</div>
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
	    		<Link key={key} href={item.url}>
	    			<a className="no-underline text-black dark:text-white">
	    				{element}
	    			</a>
	    		</Link>
    	})}
    </div>
  );
};