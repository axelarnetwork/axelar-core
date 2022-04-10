module.exports = {
  plugins: [
    'tailwindcss/nesting',
    'tailwindcss',
    'autoprefixer',
    'postcss-import',
    'postcss-nested',
    'postcss-color-function',
    'postcss-flexbugs-fixes',
    [
      'postcss-preset-env',
      {
        autoprefixer: {
          flexbox: 'no-2009'
        },
        stage: 3,
        features: {
          'custom-properties': false
        }
      }
    ]
  ]
}