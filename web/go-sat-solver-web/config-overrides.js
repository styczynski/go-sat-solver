/* config-overrides.js */
const MonacoWebpackPlugin = require('monaco-editor-webpack-plugin');

module.exports = function override(config, env) {
    if (!config.plugins) {
        config.plugins = [];
    }

    console.log(config.module.rules);
    const extensionsToExclude = /\.(less|config|variables|overrides)$/;
    for (const rule of config.module.rules) {
        if (!rule.test && rule.oneOf) {
            for (let i = 0; i < rule.oneOf.length; i++) {
                const subrule = rule.oneOf[i];
                if (subrule && subrule.loader && subrule.loader.indexOf('file-loader') > -1) {
                    // This is file loader
                    subrule.exclude.push(extensionsToExclude);
                    continue
                }

                if (rule.oneOf[i].test && !subrule.test.map) {
                    console.log("O[", subrule.test.source, "]")
                } else if (subrule.test) {
                    for (const subruleRule of subrule.test) {
                        console.log("O>[", subruleRule.source, "]")
                    }
                }
            }
            continue
        } else if (rule.test) {
            console.log("T[", rule.test.source, "]")
        }
    }

    config.module.rules = [{
        test: /\.less$/,
        use: [{
            loader: 'style-loader',
        }, {
            loader: 'css-loader', // translates CSS into CommonJS
        }, {
            loader: 'less-loader', // compiles Less to CSS
            options: {
                lessOptions: { // If you are using less-loader@5 please spread the lessOptions to options directly
                    modifyVars: {
                        // 'primary-color': '#ffffff',
                        // 'link-color': '#1DA57A',
                        // 'border-radius-base': '2px',
                        // 'background-color': '#000000',
                        //'font-size-base': '20px',
                        'primary-color': '#FF4136',
                    },
                    javascriptEnabled: true,
                },
            },
        }],
    }, ...config.module.rules, ];
    config.plugins.push(
        new MonacoWebpackPlugin()
    );
    return config;
};