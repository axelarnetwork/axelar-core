import { useRouter } from "next/router";
import Image from "next/image";

const Logo = ({ width, height }) => (
  <Image
    src="/logo/logo.png"
    alt=""
    width={width}
    height={height}
  />
);

const TITLE_WITH_TRANSLATIONS = {
  "en-US": "Resources & Documentation",
};

const FEEDBACK_LINK_WITH_TRANSLATIONS = {
  "en-US": "Question? Give us feedback →",
};

export default {
  projectLink: "https://github.com/axelarnetwork/axelar-docs",
  docsRepositoryBase: "https://github.com/axelarnetwork/axelar-docs/blob/main/pages",
  titleSuffix: " | Axelar Network",
  search: true,
  unstable_flexsearch: true,
  floatTOC: true,
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
        <Logo width={24} height={24} />
        <span
          className="mx-2 font-extrabold hidden md:inline select-none"
          title={`Axelar Network | ${TITLE_WITH_TRANSLATIONS[locale] || TITLE_WITH_TRANSLATIONS["en-US"]}`}
        >
          Axelar Network
        </span>
      </>
    );
  },
  head: ({ title, meta }) => {
    const { route } = useRouter();
    const ogImage = meta.image;
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
            "The documentation for the Axelar network"
          }
        />
        <meta
          name="og:description"
          content={
            meta.description ||
            "The documentation for the Axelar network"
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
  footerText: ({ locale }) => {
    switch (locale) {
      default:
        return (
          <a
            href="https://axelar.network"
            target="_blank"
            rel="noopener"
            className="inline-flex items-center no-underline text-current font-semibold"
          >
            <span>© {new Date().getFullYear()} Axelar, Inc.</span>
          </a>
        );
    }
  },
  i18n: [
    { locale: "en-US", text: "English" },
  ],
};
