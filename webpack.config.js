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
const rootFaviconFiles = [
    "favicon.ico",
    "browserconfig.xml"
].map(function (filename) {
    return relativePath('assets/favicon/' + filename)
});

const fileLoader = function(outputPath, name) {
    return [
        {
            loader: 'file-loader',
            options: {
                outputPath: outputPath,
                name: name,
            }
        }
    ];
};

const responsive = function(jsonOutputPath, imageName) {
    return [
        {
            loader: 'file-loader',
            options: {
                outputPath: jsonOutputPath,
                name: '[name].[ext].json',
            }
        },
        'eval-loader',
        {
            loader: 'responsive-loader',
            options: {
                name: imageName,
                quality: 85, // this is default for JPEG, making it explicit
                adapter: require('responsive-loader/sharp'),
                sizes: [325, 750, 1500, 3000, 6000]
            }
        }
    ];
};

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
                        options: { sourceMap: !isProduction }
                    }
                ]),
            },
            {
                test: /\.css$/,
                use: cssLoaders
            },
            {
                exclude: rootFaviconFiles,
                include: relativePath('assets/favicon'),
                use: fileLoader('favicon/', '[name].[ext]')
            },
            {
                include: rootFaviconFiles,
                use: fileLoader('../', '[name].[ext]')
            },
            {
                test: imageTest,
                include: relativePath('assets/images'),
                use: fileLoader('images/', filename + '.[ext]')
            },
            {
                // load what responsive-loader can't
                test: /\.(gif)$/,
                include: relativePath('content'),
                use: fileLoader(contentImagesPath, filename + '.[ext]')
            },
            {
                // no support for gif
                test: /\.(png|jpg)$/,
                include: relativePath('content'),
                use: responsive('content/responsive/', contentImagesPath + filename + '-[width].[ext]')
            },
        ],
    },

    plugins: [
        // lossless compression, responsive-loader will do a quality change on JPEG to 85 quality
        new ImageminPlugin({
            test: /\.(png|gif|svg)$/,
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