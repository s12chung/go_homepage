const path = require('path');
const relativePath = function(p) { return path.resolve(__dirname, p); };

const isProduction = process.env.NODE_ENV === "production";
const filenameF = function() { return isProduction ? '[name]-[hash]' : '[name]'; };

const DefStatic = require('webpack-def-static')(__dirname, filenameF);
const DefImages = require('webpack-def-images')(__dirname, filenameF);
const DefSass = require('webpack-def-sass')(__dirname, filenameF);

module.exports = {
    mode: isProduction ? "production" : "development",

    entry: {
        main: relativePath('assets/js/main.js'),
        vendor: relativePath('assets/js/vendor.js'),
    },
    output: {
        path: relativePath('generated/assets'),
        filename: filenameF() + '.js',
    },

    module: {
        rules: DefSass.sassRules()
            .concat(DefImages.faviconRules('assets/favicon'))
            .concat(DefImages.responsiveRules(relativePath('assets/images'), 'images/'))
            .concat(DefImages.responsiveRules(relativePath('content'), 'content/images/'))
    },

    plugins: DefStatic.allPlugins([{ test: DefSass.hardSourceExcludeTest }])
        .concat(DefImages.optimizationPlugins())
        .concat(DefSass.extractPlugins())
};