const typescript = require('@rollup/plugin-typescript')

exports.default = {
    input: "src/uncompress.ts",
    output: {
        file: "dist/dogeicon-un.js",
        format: "iife",
        name: "DogeIcon",
        banner: "// DogeIcon Compression Library Version 1.0\n// Copyright (c) 2024 DogeOrg. MIT License.\n",
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
