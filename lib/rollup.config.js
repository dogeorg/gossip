const typescript = require('@rollup/plugin-typescript')

exports.default = {
    input: "dogeicon.ts",
    output: {
        file: "dogeicon.js",
        format: "iife",
        name: "DogeIcon",
        sourcemap: true,
    },
    plugins: [
        typescript({
            "compilerOptions": {
                "target": "es5",
                "module": "esnext",
            }
        }),
    ],
}
