const path = require('path');

const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const ManifestPlugin = require('webpack-manifest-plugin');
const HardSourceWebpackPlugin = require('hard-source-webpack-plugin');

module.exports = {
    mode: 'development',

    module: {
        rules: [{
            test: /\.scss$/,
            use: [
                MiniCssExtractPlugin.loader,
                "css-loader",
                "sass-loader"
            ]
        }],
    },

    entry: path.resolve(__dirname, 'assets/js/main.js'),
    output: {
        path: path.resolve(__dirname, 'generated/assets'),
    },

    plugins: [
        new MiniCssExtractPlugin(),
        new ManifestPlugin(),
        new HardSourceWebpackPlugin(),
        new HardSourceWebpackPlugin.ExcludeModulePlugin([
            {
                // does not emit for repeated builds
                test: /mini-css-extract-plugin[\\/]dist[\\/]loader/,
            },
        ]),
    ]
};