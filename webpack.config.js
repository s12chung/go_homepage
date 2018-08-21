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
const faviconRules = function (faviconFilesPath) {
    let rootFaviconFiles = [
        "favicon.ico",
        "browserconfig.xml"
    ].map(function (filename) {
        return relativePath(faviconFilesPath + "/" + filename)
    });

    return  [
        {
            exclude: rootFaviconFiles,
            include: relativePath(faviconFilesPath),
            use: fileLoader('favicon/', '[name].[ext]')
        },
        {
            include: rootFaviconFiles,
            use: fileLoader('../', '[name].[ext]')
        }
    ];
};

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

const responsiveExt = /\.(png|jpg)$/;
const nonResponsiveExt = /\.(gif|svg)$/;
const responsiveRules = function(include, outputPath, filenameWithoutExt) {
    return [
        {
            test: nonResponsiveExt,
            include: include,
            use: fileLoader(outputPath, filenameWithoutExt + '.[ext]')
        },
        {
            test: responsiveExt,
            include: include,
            use: responsive(outputPath + "responsive/", outputPath + filenameWithoutExt + '-[width].[ext]')
        }
    ]
};

const responsive = function(jsonOutputPath, imageName) {
    return [
        {
            loader: 'file-loader',
            options: {
                outputPath: jsonOutputPath,
                name: '[name.[ext].json',
            }
        },
        'webpack-stringify-loader',
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
            .concat(faviconRules('assets/favicon'))
            .concat(responsiveRules(relativePath('assets/images'), 'images/', filename))
            .concat(responsiveRules(relativePath('content'), 'content/images/', filename))
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