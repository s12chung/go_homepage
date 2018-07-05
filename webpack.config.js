const path = require('path');

const isProduction = process.env.NODE_ENV === "production";

const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const ManifestPlugin = require('webpack-manifest-plugin');
const HardSourceWebpackPlugin = require('hard-source-webpack-plugin');

module.exports = {
    mode: isProduction ? "production" : "development",

    module: {
        rules: [{
            test: /\.scss$/,
            use: [
                MiniCssExtractPlugin.loader,
                {
                    loader: 'css-loader',
                    options: {
                        minimize: isProduction,
                        sourceMap: !isProduction,
                    }
                },
                {
                    loader: 'sass-loader',
                    options: {
                        sourceMap: !isProduction,
                    }
                },
            ]
        }],
    },

    entry: path.resolve(__dirname, 'assets/js/main.js'),
    output: {
        path: path.resolve(__dirname, 'generated/assets'),
        filename: isProduction ? '[name]-[hash].js' : '[name].js'
    },

    plugins: [
        new MiniCssExtractPlugin({
            filename: isProduction ? '[name]-[hash].css' : '[name].css',
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
    ]
};