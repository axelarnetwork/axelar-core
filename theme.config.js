import { useRouter } from "next/router";
import Link from "next/link";
import Image from "next/image";

const Logo = ({
  theme,
  width,
  height,
}) => (
  <>
    <div className="flex dark:hidden items-center">
      <Image
        src="/logo/logo.png"
        alt=""
        width={width}
        height={height}
      />
    </div>
    <div className="hidden dark:flex items-center">
      <Image
        src="/logo/logo_dark.png"
        alt=""
        width={width}
        height={height}
      />
    </div>
  </>
);

const TITLE_WITH_TRANSLATIONS = {
  "en-US": "Resources & Documentation",
};

const FEEDBACK_LINK_WITH_TRANSLATIONS = {
  "en-US": "Question? Give us feedback →",
};

export default {
  projectLink: "https://github.com/axelarnetwork",
  docsRepositoryBase: "https://github.com/axelarnetwork/axelar-docs/blob/main/pages",
  titleSuffix: " | Axelar Network",
  search: true,
  unstable_flexsearch: true,
  floatTOC: true,
  defaultMenuCollapsed: true,
  feedbackLink: () => {
    const { locale } = useRouter();
    return (
      FEEDBACK_LINK_WITH_TRANSLATIONS[locale] ||
      FEEDBACK_LINK_WITH_TRANSLATIONS["en-US"]
    );
  },
  feedbackLabels: "feedback",
  logo: () => {
    const { locale } = useRouter();
    return (
      <>
        <Logo
          width={24}
          height={24}
        />
        <span
          title={`Axelar Network | ${TITLE_WITH_TRANSLATIONS[locale] || TITLE_WITH_TRANSLATIONS["en-US"]}`}
          className="hidden md:inline select-none font-extrabold mx-2"
        >
          Axelar Network
        </span>
        <span className="hidden md:inline select-none text-gray-500 dark:text-gray-400 mx-1">
          Documentation
        </span>
      </>
    );
  },
  head: ({
    title,
    meta,
  }) => {
    const { route } = useRouter();
    const ogImage = meta.image || "/images/og.png";
    return (
      <>
        {/* Favicons, meta */}
        <link
          rel="apple-touch-icon"
          sizes="180x180"
          href="/favicon/apple-touch-icon.png"
        />
        <link
          rel="icon"
          type="image/png"
          sizes="32x32"
          href="/favicon/favicon-32x32.png"
        />
        <link
          rel="icon"
          type="image/png"
          sizes="16x16"
          href="/favicon/favicon-16x16.png"
        />
        <link rel="manifest" href="/favicon/site.webmanifest" />
        <link
          rel="mask-icon"
          href="/favicon/safari-pinned-tab.svg"
          color="#000000"
        />
        <meta name="msapplication-TileColor" content="#ffffff" />
        <meta httpEquiv="Content-Language" content="en" />
        <meta
          name="description"
          content={
            meta.description ||
            "Axelar Network Resources & Documentation"
          }
        />
        <meta
          name="og:description"
          content={
            meta.description ||
            "Axelar Network Resources & Documentation"
          }
        />
        <meta name="twitter:card" content="summary_large_image" />
        <meta name="twitter:site" content="@axelarcore" />
        <meta name="twitter:image" content={ogImage} />
        <meta
          name="og:title"
          content={
            title ? title + " | Axelar Network" : "Axelar Network | Documentation"
          }
        />
        <meta name="og:image" content={ogImage} />
        <meta name="apple-mobile-web-app-title" content="Axelar Network" />
      </>
    );
  },
  footerEditLink: ({ locale }) => {
    switch (locale) {
      default:
        return "Edit this page on GitHub →";
    }
  },
  footerText: (
    <div className="space-y-8 mt-4">
      <div className="grid grid-flow-row grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
        <div className="space-y-2 mx-4">
          <div>
            <div className="flex dark:hidden items-center">
              <Image
                src="/logo/logo_full.png"
                alt=""
                width={270}
                height={60}
              />
            </div>
            <div className="hidden dark:flex items-center">
              <Image
                src="/logo/logo_full_dark.png"
                alt=""
                width={270}
                height={60}
              />
            </div>
          </div>
          <div className="text-lg">
            Empowering developers to build scalable, multichain dApps
          </div>
        </div>
        <div className="md:col-span-2 grid grid-flow-row grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-4">
          <div className="space-y-4 mx-4">
            <div className="text-black dark:text-white text-lg font-bold">
              Learn
            </div>
            <div className="flex flex-col space-y-3">
              <a
                href="https://axelar.network"
                target="_blank"
                rel="noopener noreferrer"
                className="no-underline text-current"
              >
                Website
              </a>
              <a
                href="https://axelar.network/wp-content/uploads/2021/07/axelar_whitepaper.pdf"
                target="_blank"
                rel="noopener noreferrer"
                className="no-underline text-current"
              >
                Whitepaper
              </a>
              <a
                href="https://axelar.network/blog"
                target="_blank"
                rel="noopener noreferrer"
                className="no-underline text-current"
              >
                Blog
              </a>
              <a
                href="https://github.com/axelarnetwork"
                target="_blank"
                rel="noopener noreferrer"
                className="no-underline text-current"
              >
                Github
              </a>
            </div>
          </div>
          <div className="space-y-4 mx-4">
            <div className="text-black dark:text-white text-lg font-bold">
              Community
            </div>
            <div className="flex flex-col space-y-3">
              <a
                href="https://twitter.com/axelarcore"
                target="_blank"
                rel="noopener noreferrer"
                className="no-underline text-current"
              >
                Twitter
              </a>
              <a
                href="https://discord.gg/aRZ3Ra6f7D"
                target="_blank"
                rel="noopener noreferrer"
                className="no-underline text-current"
              >
                Discord
              </a>
              <a
                href="https://t.me/axelarcommunity"
                target="_blank"
                rel="noopener noreferrer"
                className="no-underline text-current"
              >
                Telegram
              </a>
              <a
                href="https://community.axelar.network"
                target="_blank"
                rel="noopener noreferrer"
                className="no-underline text-current"
              >
                Discussion
              </a>
            </div>
          </div>
          <div className="space-y-4 mx-4">
            <div className="text-black dark:text-white text-lg font-bold">
              Others
            </div>
            <div className="flex flex-col space-y-3">
              <Link href="/bug-bounty">
                <a className="no-underline text-current">
                  Bug Bounty
                </a>
              </Link>
              <Link href="/terms-of-use">
                <a className="no-underline text-current">
                  Terms of Use
                </a>
              </Link>
            </div>
          </div>
        </div>
      </div>
      <div className="text-center">
        <a
          href="https://axelar.network"
          target="_blank"
          rel="noopener"
          className="inline-flex items-center no-underline text-current font-semibold"
        >
          <span className="whitespace-nowrap">
            © {new Date().getFullYear()} Axelar Network. All Rights Reserved.
          </span>
        </a>
      </div>
    </div>
  ),
};