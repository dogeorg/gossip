// DogeIcon Compression Library Version 1.0
// Copyright (c) 2024 DogeOrg. MIT License.

var DogeIcon = (function (exports) {
    'use strict';

    var Kr = 0.2126;
    var Kg = 0.7152;
    var Kb = 0.0722;
    var CbR = -0.1146;
    var CbG = -0.3854;
    var CbB = 0.5;
    var CrR = 0.5;
    var CrG = -0.4542;
    var CrB = -0.0458;
    // decoding:
    var RCr = 1.5748;
    var GCb = -0.1873;
    var GCr = -0.4681;
    var BCb = 1.8556;
    // scaling
    var scaleY = 0.1254; // as per CbCr // was: [8,240] -> [0,31]: 237+ -> 31; 12+ -> 1 // 0.1334
    var biasY = 2.0;
    var unscaleY = 8.0; // was 7.3 // [0,31] -> [14,240]
    var unbiasY = 0.0;
    var scaleCbCr = 0.1254; // [0,255] -> [0,31]: 248+ -> 31; 8+ -> 1
    var unscaleCbCr = 8.0; // [0,31] -> [0,248]; 31 -> 248;  1 -> 8
    var biasCbCr = 127.5 + 2; // [-123.5,131.5] -> [0.0,255.0] (+4 round-to-nearest)
    var unbiasCbCr = -127.984; // 16 -> 0
    function Y(Y) {
        var s = ((Y + biasY) * scaleY) | 0;
        if (s < 0)
            return 0;
        if (s > 31)
            return 31;
        return s;
    }
    function unY(Y) {
        return (Y + unbiasY) * unscaleY;
    }
    function diff(x, y) {
        return x < y ? y - x : x - y;
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

    // Indices into YTile array:
    var topoMap = [
        // flat interpolation:
        [
            [
                // '/' diagonal
                0, 0, // 0/
                0, 3, // /3
            ],
            [
                // '\' diagonal
                2, 1, // \1
                2, 2, // 2\
            ],
            [
                // '-' horizontal
                0, 0, // 00 horizontal
                2, 2, // 22 no blending
            ],
            [
                // '|' vertical
                0, 1, // 01 vertical
                0, 1, // 01 no blending
            ],
        ],
        // linear interpolation:
        [
            [
                // '/' diagonal
                0, 4, // 0/
                4, 3, // /3 gradient
            ],
            [
                // '\' diagonal
                5, 1, // \1
                2, 5, // 2\ gradient
            ],
            [
                // '-' horizontal
                6, 6, // 01 horizontal
                7, 7, // 23 centre samples 0+1 & 2+3
            ],
            [
                // '|' vertical
                8, 9, // 01 vertical
                8, 9, // 23 centre samples 0+2 & 1+3
            ],
        ],
    ];
    // func cmax(Cb0, Cr0, Cb1, Cr1 float32) (Cb, Cr float32) {
    // 	if math.Abs(float64(Cb0)) > math.Abs(float64(Cb1)) && math.Abs(float64(Cr0)) > math.Abs(float64(Cr1)) {
    // 		return Cb0, Cr0
    // 	}
    // 	return Cb1, Cr1
    // }
    function compress1(rgb, style, components, options) {
        // encode flat or linear style
        var topMap = topoMap[style & 1];
        var comp = new Uint8Array(1585);
        // encoding state
        var Yacc = 0 | 0;
        var Ybit = 10;
        var compY = 1;
        comp[0] = style;
        // for each 2x2 tile:
        var stride = 48 * components;
        for (var y = 0; y < 48; y += 2) {
            var row = y * stride;
            for (var x = 0; x < 48; x += 2) {
                var r1 = row + (x * components);
                var r2 = r1 + components;
                var r3 = row + stride + (x * components);
                var r4 = r3 + components;
                // assemble RGB tile values for topology calc
                var tile = {
                    tlR: rgb[r1], tlG: rgb[r1 + 1], tlB: rgb[r1 + 2], // TL
                    trR: rgb[r2], trG: rgb[r2 + 1], trB: rgb[r2 + 2], // TR
                    blR: rgb[r3], blG: rgb[r3 + 1], blB: rgb[r3 + 2], // BL
                    brR: rgb[r4], brG: rgb[r4 + 1], brB: rgb[r4 + 2], // BR
                };
                // 1. convert 2x2 tile to YCbCr
                var R0 = rgb[r1];
                var G0 = rgb[r1 + 1];
                var B0 = rgb[r1 + 2];
                var Y0 = R0 * Kr + G0 * Kg + B0 * Kb;
                var Cb0 = R0 * CbR + G0 * CbG + B0 * CbB;
                var Cr0 = R0 * CrR + G0 * CrG + B0 * CrB;
                var R1 = rgb[r2];
                var G1 = rgb[r2 + 1];
                var B1 = rgb[r2 + 2];
                var Y1 = R1 * Kr + G1 * Kg + B1 * Kb;
                var Cb1 = R1 * CbR + G1 * CbG + B1 * CbB;
                var Cr1 = R1 * CrR + G1 * CrG + B1 * CrB;
                // Cb0, Cr0 = cmax(Cb0, Cr0, Cb1, Cr1)
                var R2 = rgb[r3];
                var G2 = rgb[r3 + 1];
                var B2 = rgb[r3 + 2];
                var Y2 = R2 * Kr + G2 * Kg + B2 * Kb;
                var Cb2 = R2 * CbR + G2 * CbG + B2 * CbB;
                var Cr2 = R2 * CrR + G2 * CrG + B2 * CrB;
                // Cb0, Cr0 = cmax(Cb0, Cr0, Cb1, Cr1)
                var R3 = rgb[r4];
                var G3 = rgb[r4 + 1];
                var B3 = rgb[r4 + 2];
                var Y3 = R3 * Kr + G3 * Kg + B3 * Kb;
                var Cb3 = R3 * CbR + G3 * CbG + B3 * CbB;
                var Cr3 = R3 * CrR + G3 * CrG + B3 * CrB;
                // Cb0, Cr0 = cmax(Cb0, Cr0, Cb1, Cr1)
                if ((options & 24) == 8) {
                    // average 4 chroma samples
                    Cb0 = (Cb0 + Cb1 + Cb2 + Cb3) / 4.0;
                    Cr0 = (Cr0 + Cr1 + Cr2 + Cr3) / 4.0;
                }
                else if ((options & 24) == 16) {
                    // average 2 horizontal chroma samples
                    Cb0 = (Cb0 + Cb1) / 2.0;
                    Cr0 = (Cr0 + Cr1) / 2.0;
                }
                else if ((options & 24) == 24) {
                    // weighted average (using intensity as weight)
                    var b = 1 / (1021 - (Y0 + Y1 + Y2 + Y3));
                    var wx = b * ((255 - Y0) * 0 + (255 - Y1) * 1 + (255 - Y2) * 0 + (255 - Y3) * 1);
                    var wy = b * ((255 - Y0) * 0 + (255 - Y1) * 0 + (255 - Y2) * 1 + (255 - Y3) * 1);
                    Cb0 = (Cb0 * wx * wy) + (Cb1 * (1 - wx) * wy) +
                        (Cb2 * wx * (1 - wy)) + (Cb3 * (1 - wx) * (1 - wx));
                    Cr0 = (Cr0 * wx * wy) + (Cr1 * (1 - wx) * wy) +
                        (Cr2 * wx * (1 - wy)) + (Cr3 * (1 - wx) * (1 - wx));
                }
                // compressed CbCr values (quantized)
                var CbQ = ((Cb0 + biasCbCr) * scaleCbCr);
                var CrQ = ((Cr0 + biasCbCr) * scaleCbCr);
                // clamp out of range
                if (CbQ > 31)
                    CbQ = 31;
                if (CrQ > 31)
                    CrQ = 31;
                // uncompressed CbCr values
                var Cb = CbQ * unscaleCbCr + unbiasCbCr;
                var Cr = CrQ * unscaleCbCr + unbiasCbCr;
                var Red = Cr * RCr;
                var Green = Cr * GCr + Cb * GCb;
                var Blue = Cb * BCb;
                // compressed Y values (quantized)
                var Y0Q = Y(Y0);
                var Y1Q = Y(Y1);
                var Y2Q = Y(Y2);
                var Y3Q = Y(Y3);
                var Ys = [
                    Y0Q, // top-left
                    Y1Q, // top-right
                    Y2Q, // bottom-left
                    Y3Q, // bottom-right
                    (Y0Q + Y3Q) >> 1, // '/' diagonal lerp
                    (Y1Q + Y2Q) >> 1, // '\' diagonal lerp
                    (Y0Q + Y1Q) >> 1, // '-' horizontal lerp
                    (Y2Q + Y3Q) >> 1, // '-' horizontal lerp
                    (Y0Q + Y2Q) >> 1, // '|' vertical lerp
                    (Y1Q + Y3Q) >> 1, // '|' vertical lerp
                ];
                // 2. choose topology to minimise error
                var minsad = 4096 | 0; // > 255*12
                var topology = 0 | 0;
                // var sads [4]uint
                for (var q = 0 | 0; q < 4; q++) {
                    var tmap = topMap[q];
                    var Ytl = unY(Ys[tmap[0]]);
                    var tl = diff(tile.tlR, (Ytl + Red) | 0) +
                        diff(tile.tlG, (Ytl + Green) | 0) +
                        diff(tile.tlB, (Ytl + Blue) | 0);
                    var Ytr = unY(Ys[tmap[1]]);
                    var tr = diff(tile.trR, (Ytr + Red) | 0) +
                        diff(tile.trG, (Ytr + Green) | 0) +
                        diff(tile.trB, (Ytr + Blue) | 0);
                    var Ybl = unY(Ys[tmap[2]]);
                    var bl = diff(tile.blR, (Ybl + Red) | 0) +
                        diff(tile.blG, (Ybl + Green) | 0) +
                        diff(tile.blB, (Ybl + Blue) | 0);
                    var Ybr = unY(Ys[tmap[3]]);
                    var br = diff(tile.brR, (Ybr + Red) | 0) +
                        diff(tile.brG, (Ybr + Green) | 0) +
                        diff(tile.brB, (Ybr + Blue) | 0);
                    var sad = tl + tr + bl + br;
                    // sads[q] = sad
                    if (sad < minsad) {
                        minsad = sad;
                        topology = q;
                    }
                }
                // log.Printf("SAD %v %d :: %v %v %v %v :: %v", y, x, sads[0], sads[1], sads[2], sads[3], topology)
                // 3. compute encoded Y components
                var Y0bits = 0, Y1bits = 0;
                switch (topology) {
                    case 0:
                        Y0bits = Ys[topMap[0][0]];
                        Y1bits = Ys[topMap[0][3]];
                    case 1:
                        Y0bits = Ys[topMap[1][1]];
                        Y1bits = Ys[topMap[1][2]];
                    case 2:
                        Y0bits = Ys[topMap[2][0]];
                        Y1bits = Ys[topMap[2][2]];
                    case 3:
                        Y0bits = Ys[topMap[3][0]];
                        Y1bits = Ys[topMap[3][1]];
                }
                // 4. encode compressed values (22 bits)
                Yacc |= ((Y0bits << 17) | (Y1bits << 12) | (CbQ << 7) | (CrQ << 2) | topology) << Ybit;
                if (Ybit == 10) {
                    // @10 wrote 22 (store top 16)
                    comp[compY] = (Yacc >> 24);
                    comp[compY + 1] = (Yacc >> 16);
                    compY += 2;
                    Yacc = Yacc << 16; // keep 6 bits (move to top)
                    Ybit = 4; // next 22 bits at bit 4
                }
                else {
                    // @4,6,8 wrote 22 (store top 24)
                    // @4+22+6/4; @6+22+4/2 @8+22+2/0
                    comp[compY] = (Yacc >> 24);
                    comp[compY + 1] = (Yacc >> 16);
                    comp[compY + 2] = (Yacc >> 8);
                    compY += 3;
                    Yacc = Yacc << 24; // keep 4,2,0 bits (move to top)
                    Ybit += 2; // next bits at 6 -> 8 -> 10
                }
            }
        }
        return comp;
    }

    /*
    Compress a 48x48 sRGB-8 source image to 1584-byte Y'CbCr 4:2:0 `DogeIcon` (23% original size)

    Encoded as 24x24 2x2 tiles with a 2-bit tile-topology; each tile encodes Y1 Y2 Cb Cr:
    Y=5bit Cb=5bit Cr=5bit (scaled full-range; no "headroom"; using BT.709)

    plane0 Y1Y2 (24*24*5*2/8=720) || plane1 CbCr (24*24*5*2/8=720) || plane2 topology (24*24*2/8=144) = 1584 (1.55K)

        0 = 0/  1 = \1  2 = 00  3 = 02   2 Y-samples per tile (5-bit samples packed into 10-bit plane)
            /3      2\      22      02   2-bit topology as per diagram (0=/ 1=\ 2=H 3=V)

    Bytes are filled with bits left to right (left=MSB right=LSB)
    */
    function compress(rgb, components, options) {
        if ((options & 4) != 0) {
            var style = options & 1;
            var comp = compress1(rgb, style, components, options);
            var res = uncompress(comp);
            //console.log(`MODE ${style}`)
            return { comp: comp, res: res };
        }
        // try both and use the least-difference version
        var compFlat = compress1(rgb, 0, components, options);
        var compLinear = compress1(rgb, 1, components, options);
        var resFlat = uncompress(compFlat);
        var resLinear = uncompress(compLinear);
        if (sad(rgb, resFlat) < sad(rgb, resLinear)) {
            //console.log("MODE 0")
            return { comp: compFlat, res: resFlat };
        }
        //console.log("MODE 1")
        return { comp: compLinear, res: resLinear };
    }
    // Sum of Absolute Difference between 48x48 images.
    function sad(rgb, res) {
        var sum = 0.0; // safe up to 2369x2369 image
        for (var x = 0; x < 48 * 48 * 3; x += 3) {
            //sum += diff(uint(rgb[x]), uint(res[x]))
            var R0 = rgb[x];
            var G0 = rgb[x + 1];
            var B0 = rgb[x + 2];
            var Y0 = R0 * Kr + G0 * Kg + B0 * Kb;
            var R1 = res[x];
            var G1 = res[x + 1];
            var B1 = res[x + 2];
            var Y1 = R1 * Kr + G1 * Kg + B1 * Kb;
            sum += diff(Y0 | 0, Y1 | 0);
        }
        return sum;
    }

    exports.compress = compress;

    return exports;

})({});
