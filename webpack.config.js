const path = require('path');

const relativePath = function(p) { return path.resolve(__dirname, p); };

const isProduction = process.env.NODE_ENV === "production";

const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const ManifestPlugin = require('webpack-manifest-plugin');
const HardSourceWebpackPlugin = require('hard-source-webpack-plugin');

const ImageminPlugin = require('imagemin-webpack-plugin').default;

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
const imageTest = /\.(png|jpg|gif)$/;
const contentImagesPath = 'content/images/';

module.exports = {
    resolveLoader: {
        modules: ['node_modules', 'webpack/loaders']
    },

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
                test: imageTest,
                include: relativePath('assets/images'),
                use: [
                    {
                        loader: 'file-loader',
                        options: {
                            outputPath: 'images/',
                            name: filename + '.[ext]',
                        }
                    }
                ]
            },
            {
                // load what responsive-loader can't
                test: /\.(gif)$/,
                include: relativePath('content'),
                use: [
                    {
                        loader: 'file-loader',
                        options: {
                            outputPath: contentImagesPath,
                            name: filename + '.[ext]',
                        }
                    }
                ]
            },
            {
                // no support for gif
                test: /\.(png|jpg)$/,
                include: relativePath('content'),
                use: [
                    {
                        loader: 'file-loader',
                        options: {
                            outputPath: 'content/responsive/',
                            name: '[name].[ext].json',
                        }
                    },
                    'eval-loader',
                    {
                        loader: 'responsive-loader',
                        options: {
                            name: contentImagesPath + filename + '-[width].[ext]',
                            quality: 85, // this is default for JPEG, making it explicit
                            adapter: require('responsive-loader/sharp'),
                            sizes: [325, 750, 1500, 3000, 6000]
                        }
                    }
                ]
            },
        ],
    },

    plugins: [
        // lossless compression, responsive-loader will do a quality change on JPEG to 85 quality
        new ImageminPlugin({
            test: /\.(png|gif)$/,
            cacheFolder: relativePath('node_modules/.cache/imagemin'),
            jpegtran: null
        }),
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
    ]
};