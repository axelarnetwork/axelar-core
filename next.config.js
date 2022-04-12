const withNextra = require("nextra")({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.js",
  unstable_flexsearch: true,
  unstable_staticImage: true,
});

module.exports = withNextra({
  i18n: {
    locales: ["en-US"],
    defaultLocale: "en-US",
  },
  redirects: () => {
    return [
      {
        source: "/node",
        destination: "/node/join",
        statusCode: 301,
      },
      {
        source: "/validator",
        destination: "/validator/setup",
        statusCode: 301,
      },
      {
        source: "/validator/setup",
        destination: "/validator/setup/overview",
        statusCode: 301,
      },
      {
        source: "/validator/external-chains",
        destination: "/validator/external-chains/overview",
        statusCode: 301,
      },
      {
        source: "/validator/troubleshoot",
        destination: "/validator/troubleshoot/startup",
        statusCode: 301,
      },
      {
        source: "/resources/mainnet-releases",
        destination: "/resources/mainnet",
        statusCode: 301,
      },
      {
        source: "/resources/testnet-releases",
        destination: "/resources/testnet",
        statusCode: 301,
      },
      {
        source: "/releases/:path",
        destination: "/resources/:path",
        statusCode: 301,
      },
      {
        source: "/user/:path",
        destination: "/resources/:path",
        statusCode: 301,
      },
      {
        source: "/roles/:path",
        destination: "/:path",
        statusCode: 301,
      },
    ];
  },
});
