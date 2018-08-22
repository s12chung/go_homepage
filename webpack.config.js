const path = require('path');

const relativePath = function(p) { return path.resolve(__dirname, p); };

const isProduction = process.env.NODE_ENV === "production";

const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const ManifestPlugin = require('webpack-manifest-plugin');
const HardSourceWebpackPlugin = require('hard-source-webpack-plugin');

const DefImages = require('webpack-def-images')(__dirname);

const cssLoaders = [
    MiniCssExtractPlugin.loader,
    {
        loader: 'css-loader',
        options: {
            minimize: isProduction,
            sourceMap: !isProduction,
        }
    }
];

const filename = isProduction ? '[name]-[hash]' : '[name]';
module.exports = {
    mode: isProduction ? "production" : "development",

    entry: {
        main: relativePath('assets/js/main.js'),
        vendor: relativePath('assets/js/vendor.js'),
    },
    output: {
        path: relativePath('generated/assets'),
        filename: filename + '.js',
    },

    module: {
        rules: [
            {
                test: /\.scss$/,
                use: cssLoaders.concat([
                    { loader: 'sass-loader', options: { sourceMap: !isProduction } }
                ]),
            },
            {
                test: /\.css$/,
                use: cssLoaders
            },
        ]
            .concat(DefImages.faviconRules('assets/favicon'))
            .concat(DefImages.responsiveRules(relativePath('assets/images'), 'images/', filename))
            .concat(DefImages.responsiveRules(relativePath('content'), 'content/images/', filename))
    },

    plugins: [
        new MiniCssExtractPlugin({
            filename: filename + '.css',
            chunkFilename: isProduction ? '[id]-[hash].css' : '[id].css',
        }),
        new ManifestPlugin(),
        new HardSourceWebpackPlugin(),
        new HardSourceWebpackPlugin.ExcludeModulePlugin([
            {
                // does not emit for repeated builds
                test: /mini-css-extract-plugin[\\/]dist[\\/]loader/,
            },
        ]),
    ].concat(DefImages.optimizationPlugins())
};