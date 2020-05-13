const path = require('path');
const isProd = process.argv.indexOf('--debug') === -1;

module.exports = {
    mode: 'production',
    entry: './src/plugin.js',
    output: {
        filename: 'plugin-gravatar.js',
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                use: [{loader: 'exports-loader'}, {loader: 'babel-loader'}],
                include: [
                    path.join(__dirname, 'src'),
                ]
            },
        ]
    },
    devtool: isProd ? '' : 'source-map',
    devServer: {
        filename: 'plugin-gravatar.js',
        contentBase: path.join(__dirname, "dist"),
        compress: true,
        port: 9000
    }
};