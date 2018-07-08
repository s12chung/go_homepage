const path = require('path');

const relativePath = function(p) { return path.resolve(__dirname, p); };

const isProduction = process.env.NODE_ENV === "production";

const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const ManifestPlugin = require('webpack-manifest-plugin');
const HardSourceWebpackPlugin = require('hard-source-webpack-plugin');

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

module.exports = {
    mode: isProduction ? "production" : "development",

    entry: {
        main: relativePath('assets/js/main.js'),
        vendor: relativePath('assets/js/vendor.js'),
    },
    output: {
        path: relativePath('generated/assets'),
        filename: isProduction ? '[name]-[hash].js' : '[name].js',
    },

    module: {
        rules: [
            {
                test: /\.scss$/,
                use: cssLoaders.concat([
                    {
                        loader: 'sass-loader',
                        options: {
                            sourceMap: !isProduction,
                        }
                    }
                ]),
            },
            {
                test: /\.css$/,
                use: cssLoaders
            },
            {
                test: /\.(png|jpg|gif)$/,
                use: [
                    {
                        loader: 'file-loader',
                        options: {
                            outputPath: 'images/',
                            name: isProduction ? '[name]-[hash].[ext]' : '[name].[ext]',
                        }
                    }
                ]
            }
        ],
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