const isProduction = process.env.NODE_ENV === "production";
const filenameF = function() { return isProduction ? '[name]-[hash]' : '[name]'; };

const defaults = require('gostatic-webpack')(__dirname, filenameF);
const imageGenerator = require('gostatic-webpack-images')(__dirname, filenameF);

const relativePath = function(p) { return require('path').resolve(__dirname, p); };

module.exports = {
    mode: isProduction ? "production" : "development",

    entry: Object.assign(defaults.entry(), {
    }),
    output: Object.assign(defaults.output(), {
    }),

    module: {
        rules: defaults.allRules().concat(
            imageGenerator.responsiveRules(relativePath('content'), 'content/images/')
        )
    },

    plugins: defaults.allPlugins().concat([
    ])
};