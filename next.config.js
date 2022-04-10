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
        source: "/node",
        destination: "/node/join",
        statusCode: 302,
      },
      {
        source: "/validator",
        destination: "/validator/setup",
        statusCode: 302,
      },
      {
        source: "/validator/setup",
        destination: "/validator/setup/overview",
        statusCode: 302,
      },
      {
        source: "/validator/external-chains",
        destination: "/validator/external-chains/overview",
        statusCode: 302,
      },
    ];
  },
});
