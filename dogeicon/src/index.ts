import { uncompress } from "./uncompress"
import { compress1 } from "./compress"
import { diff, Kb, Kg, Kr } from "./constants"

/*
Compress a 48x48 sRGB-8 source image to 1584-byte Y'CbCr 4:2:0 `DogeIcon` (23% original size)

Encoded as 24x24 2x2 tiles with a 2-bit tile-topology; each tile encodes Y1 Y2 Cb Cr:
Y=5bit Cb=5bit Cr=5bit (scaled full-range; no "headroom"; using BT.709)

plane0 Y1Y2 (24*24*5*2/8=720) || plane1 CbCr (24*24*5*2/8=720) || plane2 topology (24*24*2/8=144) = 1584 (1.55K)

	0 = 0/  1 = \1  2 = 00  3 = 02   2 Y-samples per tile (5-bit samples packed into 10-bit plane)
	    /3      2\      22      02   2-bit topology as per diagram (0=/ 1=\ 2=H 3=V)

Bytes are filled with bits left to right (left=MSB right=LSB)
*/
export function compress(rgb: Uint8Array, components: number, options: number): {comp:Uint8Array,res:Uint8Array} {
	if ((options&4) != 0) {
		const style = options & 1
		const comp = compress1(rgb, style, components, options)
		const res = uncompress(comp)
		//console.log(`MODE ${style}`)
		return {comp, res}
	}
	// try both and use the least-difference version
	const compFlat = compress1(rgb, 0, components, options)
	const compLinear = compress1(rgb, 1, components, options)
	const resFlat = uncompress(compFlat)
	const resLinear = uncompress(compLinear)
	if (sad(rgb, resFlat) < sad(rgb, resLinear)) {
		//console.log("MODE 0")
		return {comp:compFlat, res:resFlat}
	}
	//console.log("MODE 1")
	return {comp:compLinear, res:resLinear}
}

// Sum of Absolute Difference between 48x48 images.
function sad(rgb: Uint8Array, res: Uint8Array): number {
	let sum = 0.0 // safe up to 2369x2369 image
	for (let x = 0; x < 48*48*3; x += 3) {
		//sum += diff(uint(rgb[x]), uint(res[x]))
		const R0 = rgb[x]
		const G0 = rgb[x+1]
		const B0 = rgb[x+2]
		const Y0 = R0*Kr + G0*Kg + B0*Kb
		const R1 = res[x]
		const G1 = res[x+1]
		const B1 = res[x+2]
		const Y1 = R1*Kr + G1*Kg + B1*Kb
		sum += diff(Y0|0, Y1|0)
	}
	return sum
}
