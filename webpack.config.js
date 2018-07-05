const path = require('path');

const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const ManifestPlugin = require('webpack-manifest-plugin');

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
    ]
};