// DogeIcon Compression Library Version 1.0
// Copyright (c) 2024 DogeOrg. MIT License.

var DogeIcon = (function (exports) {
    'use strict';

    // decoding:
    var RCr = 1.5748;
    var GCb = -0.1873;
    var GCr = -0.4681;
    var BCb = 1.8556;
    var unscaleY = 8.0; // was 7.3 // [0,31] -> [14,240]
    var unbiasY = 0.0;
    var unscaleCbCr = 8.0; // [0,31] -> [0,248]; 31 -> 248;  1 -> 8
    var unbiasCbCr = -127.984; // 16 -> 0
    function unY(Y) {
        return (Y + unbiasY) * unscaleY;
    }

    /*
    Uncompress a `DogeIcon` to 48x48 sRGB-8 (see Compress)
    Requires 1586 bytes (1 style + 1584 data) plus 1 padding byte for the decoder!

    Style bits:
    bit 0: interpolation: 0=pixel 1=bilinear

    flat: assign pixels the closest Y-value (proportional Y in tied-pixels, or tie-break consistently)
    pixelated: divide the tile into four equal-sized squares (proportional Y or tie-break off-axis diagonals?)
    bilinear and bicubic: in horizontal and vertical tiles: position Y at centres of point-pairs
    bilinear: calculate missing corner points bilinearly (cross-tile); then standard bilinear scaling
    bicubic: calculate missing corner points bicubically (cross-tile); then standard bicubic scaling (ideally)
    */
    function uncompress(comp) {
        // decoding state
        var rgb = new Uint8Array(48 * 48 * 3);
        var Yacc = comp[1] << 24;
        var compY = 2;
        var Ybit = 0;
        var linear = comp[0] & 1;
        // for each 2x2 tile:
        var stride = 48 * 3;
        for (var y = 0; y < 48; y += 2) {
            var row = y * stride;
            for (var x = 0; x < 48; x += 2) {
                // 4. decode compressed values (22 bits)
                if (Ybit == 0) {
                    // @8 read 16 bits + 8 = 24 (use 22 keep 2)
                    Yacc |= (comp[compY] << 16) | (comp[compY + 1] << 8);
                    compY += 2;
                    Ybit = 6; // next 22 bits at bit 6
                }
                else {
                    // @6,4,2 read 24 bits + 2,4,6 = 26,28,30 (use 22 keep 4,6,8)
                    // @6+24+2/4; @4+24+4/6; @2+24+6/8
                    Yacc |= ((comp[compY] << 16) | (comp[compY + 1] << 8) | comp[compY + 2]) << Ybit;
                    compY += 3;
                    Ybit -= 2; // next 22 bits at bit 4,2,0
                }
                var Y0 = (Yacc >> 27) & 31;
                var Y1 = (Yacc >> 22) & 31;
                var Ya = unY(Y0);
                var Yb = unY(Y1);
                var Cb = ((Yacc >> 17) & 31) * unscaleCbCr + unbiasCbCr;
                var Cr = ((Yacc >> 12) & 31) * unscaleCbCr + unbiasCbCr;
                var topology = (Yacc >> 10) & 3;
                Yacc = Yacc << 22; // keep low 2,4,6,8 bits
                // 2. interpolation
                var Ytl = 0.0, Ytr = 0.0, Ybl = 0.0, Ybr = 0.0;
                if (linear != 0) {
                    // linear interpolation
                    switch (topology) {
                        case 0: // '/' diagonal
                            Ytl = Ya; // 0
                            Ybr = Yb; // 3
                            Ytr = unY((Y0 + Y1) >> 1);
                            Ybl = Ytr;
                        case 1: // '\' diagonal
                            Ytr = Ya; // 1
                            Ybl = Yb; // 2
                            Ytl = unY((Y0 + Y1) >> 1);
                            Ybr = Ytl;
                        case 2: // '-' horizontal
                            Ytl = Ya; // 0
                            Ytr = Ytl;
                            Ybl = Yb; // 2
                            Ybr = Ybl;
                        case 3: // '|' vertical
                            Ytl = Ya; // 0
                            Ybl = Ytl;
                            Ytr = Yb; // 1
                            Ybr = Ytr;
                    }
                }
                else {
                    // flat interpolation
                    switch (topology) {
                        case 0: // '/' diagonal
                            Ytl = Ya; // 0 0
                            Ytr = Ya;
                            Ybl = Ya; // 0 3
                            Ybr = Yb;
                        case 1: // '\' diagonal
                            Ytl = Yb; // 2 1
                            Ytr = Ya;
                            Ybl = Yb; // 2 2
                            Ybr = Yb;
                        case 2: // '-' horizontal
                            Ytl = Ya; // 0 0
                            Ytr = Ya;
                            Ybl = Yb; // 2 2
                            Ybr = Yb;
                        case 3: // '|' vertical
                            Ytl = Ya; // 0 1
                            Ytr = Yb;
                            Ybl = Ya; // 0 1
                            Ybr = Yb;
                    }
                }
                // 3. generate pixels
                var Red = Cr * RCr;
                var Green = Cr * GCr + Cb * GCb;
                var Blue = Cb * BCb;
                var r1 = row + (x * 3);
                var r2 = r1 + stride;
                // top-left pixel
                rgb[r1] = clamp(Ytl + Red);
                rgb[r1 + 1] = clamp(Ytl + Green);
                rgb[r1 + 2] = clamp(Ytl + Blue);
                // top-right pixel
                rgb[r1 + 3] = clamp(Ytr + Red);
                rgb[r1 + 4] = clamp(Ytr + Green);
                rgb[r1 + 5] = clamp(Ytr + Blue);
                // bottom-left pixel
                rgb[r2] = clamp(Ybl + Red);
                rgb[r2 + 1] = clamp(Ybl + Green);
                rgb[r2 + 2] = clamp(Ybl + Blue);
                // bottom-right pixel
                rgb[r2 + 3] = clamp(Ybr + Red);
                rgb[r2 + 4] = clamp(Ybr + Green);
                rgb[r2 + 5] = clamp(Ybr + Blue);
            }
        }
        return rgb;
    }
    function clamp(x) {
        // can go out of range due to CbCr averaging
        var y = x | 0;
        if (y >= 0 && y <= 255)
            return y | 0;
        return y >= 0 ? 255 : 0;
    }

    exports.uncompress = uncompress;

    return exports;

})({});
