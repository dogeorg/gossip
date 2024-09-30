export const Kr  = 0.2126
export const Kg  = 0.7152
export const Kb  = 0.0722
export const CbR = -0.1146
export const CbG = -0.3854
export const CbB = 0.5
export const CrR = 0.5
export const CrG = -0.4542
export const CrB = -0.0458
// decoding:
export const RCr = 1.5748
export const GCb = -0.1873
export const GCr = -0.4681
export const BCb = 1.8556
// scaling
export const scaleY      = 0.1254 // as per CbCr // was: [8,240] -> [0,31]: 237+ -> 31; 12+ -> 1 // 0.1334
export const biasY       = 2.0
export const unscaleY    = 8.0 // was 7.3 // [0,31] -> [14,240]
export const unbiasY     = 0.0
export const scaleCbCr   = 0.1254    // [0,255] -> [0,31]: 248+ -> 31; 8+ -> 1
export const unscaleCbCr = 8.0       // [0,31] -> [0,248]; 31 -> 248;  1 -> 8
export const biasCbCr    = 127.5 + 2 // [-123.5,131.5] -> [0.0,255.0] (+4 round-to-nearest)
export const unbiasCbCr  = -127.984  // 16 -> 0

export function Y(Y: number): number {
	let s = ((Y + biasY) * scaleY)|0
	if (s < 0) return 0
	if (s > 31) return 31
	return s
}

export function unY(Y: number): number {
	return (Y + unbiasY) * unscaleY
}

export function diff(x, y: number): number {
	return x < y ? y - x : x - y
}
