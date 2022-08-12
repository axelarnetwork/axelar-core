const withNextra = require("nextra")({
  theme: "nextra-theme-docs",
  themeConfig: "./theme.config.js",
  unstable_flexsearch: true,
  unstable_staticImage: true,
});

module.exports = withNextra({
  redirects: () => {
    return [
      {
        source: "/node",
        destination: "/node/config-node",
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
        source: "/releases/:slug*",
        destination: "/resources/:slug*",
        statusCode: 301,
      },
      {
        source: "/user/:slug*",
        destination: "/resources/:slug*",
        statusCode: 301,
      },
      {
        source: "/roles/:slug*",
        destination: "/:slug*",
        statusCode: 301,
      },
      {
        source: "/dev",
        destination: "/dev/intro",
        statusCode: 301,
      },
      {
        source: "/intro",
        destination: "/learn",
        statusCode: 301,
      },
      {
        source: "/dev/sdk",
        destination: "/learn/sdk",
        statusCode: 301,
      },
      {
        source: "/dev/sdk/:slug*",
        destination: "/learn/sdk/:slug*",
        statusCode: 301,
      },
      {
        source: "/dev/cli",
        destination: "/learn/cli",
        statusCode: 301,
      },
      {
        source: "/dev/cli/:slug*",
        destination: "/learn/cli/:slug*",
        statusCode: 301,
      },
      {
        source: "/resources/supported",
        destination: "/dev/chain-names",
        statusCode: 301,
      },
      {
        source: "/resources/weth",
        destination: "/resources/wrapped-tokens",
        statusCode: 301,
      },
      {
        source: "/dev/gmp/examples",
        destination: "/dev/build/5-min-starter-examples",
        statusCode: 301,
      },
      {
        source: "/dev/gmp/overview",
        destination: "/dev/gmp-overview",
        statusCode: 301,
      },
    ];
  },
});
