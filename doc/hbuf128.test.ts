import {test} from "node:test";
import assert from "node:assert";
import * as fs from "node:fs";
import {
    DataDescriptor,
    Descriptor,
    FieldDesc,
    HData,
    Marshal,
    Unmarshal,
    NewDataDescriptor,
    NewInt8Descriptor,
    NewInt16Descriptor,
    NewInt32Descriptor,
    NewInt64Descriptor,
    NewUint8Descriptor,
    NewUint16Descriptor,
    NewUint32Descriptor,
    NewUint64Descriptor,
    NewFloatDescriptor,
    NewDoubleDescriptor,
    NewBoolDescriptor,
    NewStringDescriptor,
    NewBytesDescriptor,
    NewTimeDescriptor,
    NewDecimalDescriptor,
    NewListDescriptor,
    NewMapDescriptor,
    NewRefDescriptor,
} from "../src";

// ===================== HBufSubTestTs（对应 Go HBufSubTest: A1 int64, A2 *int64） =====================

class HBufSubTestTs implements HData {
    a1: bigint = 0n
    a2: bigint | null = null

    descriptors(): Descriptor {
        return hbSubDesc
    }
}

const hbSubDesc: DataDescriptor = NewDataDescriptor(
    () => new HBufSubTestTs(),
    new Map<number, FieldDesc>([
        [1, {desc: NewInt64Descriptor(), get: (o: HBufSubTestTs) => o.a1, set: (o: HBufSubTestTs, v) => o.a1 = v}],
        [2, {desc: NewInt64Descriptor(true), get: (o: HBufSubTestTs) => o.a2, set: (o: HBufSubTestTs, v) => o.a2 = v}],
    ]),
)

// ===================== HBufTestTs（对应 Go HBufTest，128 字段） =====================

class HBufTestTs implements HData {

    v1: number = 0
    v2: number | null = null
    v3: number = 0
    v4: number | null = null
    v5: number = 0
    v6: number | null = null
    v7: bigint = 0n
    v8: bigint | null = null
    v9: number = 0
    v10: number | null = null
    v11: number = 0
    v12: number | null = null
    v13: number = 0
    v14: number | null = null
    v15: bigint = 0n
    v16: bigint | null = null
    v17: boolean = false
    v18: boolean | null = null
    v19: string = ""
    v20: string | null = null
    v21: Uint8Array = new Uint8Array(0)
    v22: Uint8Array | null = null
    v23: number = 0
    v24: number | null = null
    v25: number = 0
    v26: number | null = null
    v27: Date | null
    v28: Date | null = null
    v29: string = ""
    v30: string | null = null
    v31: HBufSubTestTs = new HBufSubTestTs()
    v32: HBufSubTestTs | null = null
    v33: number[] = []
    v34: (number | null)[] = []
    v35: number[] = []
    v36: (number | null)[] = []
    v37: number[] = []
    v38: (number | null)[] = []
    v39: bigint[] = []
    v40: (bigint | null)[] = []
    v41: number[] = []
    v42: (number | null)[] = []
    v43: number[] = []
    v44: (number | null)[] = []
    v45: number[] = []
    v46: (number | null)[] = []
    v47: bigint[] = []
    v48: (bigint | null)[] = []
    v49: boolean[] = []
    v50: (boolean | null)[] = []
    v51: string[] = []
    v52: (string | null)[] = []
    v53: Uint8Array[][] = []
    v54: (Uint8Array[] | null) = null
    v55: number[] = []
    v56: (number | null)[] = []
    v57: number[] = []
    v58: (number | null)[] = []
    v59: (Date | null)[] = []
    v60: (Date | null)[] = []
    v61: string[] = []
    v62: (string | null)[] = []
    v63: HBufSubTestTs[] = []
    v64: (HBufSubTestTs | null)[] = []
    v65: Map<bigint, number> = new Map()
    v66: Map<bigint, number | null> = new Map()
    v67: Map<bigint, number> = new Map()
    v68: Map<bigint, number | null> = new Map()
    v69: Map<bigint, number> = new Map()
    v70: Map<bigint, number | null> = new Map()
    v71: Map<bigint, bigint> = new Map()
    v72: Map<bigint, bigint | null> = new Map()
    v73: Map<bigint, number> = new Map()
    v74: Map<bigint, number | null> = new Map()
    v75: Map<bigint, number> = new Map()
    v76: Map<bigint, number | null> = new Map()
    v77: Map<bigint, number> = new Map()
    v78: Map<bigint, number | null> = new Map()
    v79: Map<bigint, bigint> = new Map()
    v80: Map<bigint, bigint | null> = new Map()
    v81: Map<bigint, boolean> = new Map()
    v82: Map<bigint, boolean | null> = new Map()
    v83: Map<bigint, string> = new Map()
    v84: Map<bigint, string | null> = new Map()
    v85: Map<bigint, Uint8Array[]> = new Map()
    v86: Map<bigint, Uint8Array[]> = new Map()
    v87: Map<bigint, number> = new Map()
    v88: Map<bigint, number | null> = new Map()
    v89: Map<bigint, number> = new Map()
    v90: Map<bigint, number | null> = new Map()
    v91: Map<bigint, Date | null> = new Map()
    v92: Map<bigint, Date | null> = new Map()
    v93: Map<bigint, string> = new Map()
    v94: Map<bigint, string | null> = new Map()
    v95: Map<bigint, HBufSubTestTs> = new Map()
    v96: Map<bigint, HBufSubTestTs | null> = new Map()
    v97: Map<string, number> = new Map()
    v98: Map<string, number | null> = new Map()
    v99: Map<string, number> = new Map()
    v100: Map<string, number | null> = new Map()
    v101: Map<string, number> = new Map()
    v102: Map<string, number | null> = new Map()
    v103: Map<string, bigint> = new Map()
    v104: Map<string, bigint | null> = new Map()
    v105: Map<string, number> = new Map()
    v106: Map<string, number | null> = new Map()
    v107: Map<string, number> = new Map()
    v108: Map<string, number | null> = new Map()
    v109: Map<string, number> = new Map()
    v110: Map<string, number | null> = new Map()
    v111: Map<string, bigint> = new Map()
    v112: Map<string, bigint | null> = new Map()
    v113: Map<string, boolean> = new Map()
    v114: Map<string, boolean | null> = new Map()
    v115: Map<string, string> = new Map()
    v116: Map<string, string | null> = new Map()
    v117: Map<string, Uint8Array[]> = new Map()
    v118: Map<string, Uint8Array[]> = new Map()
    v119: Map<string, number> = new Map()
    v120: Map<string, number | null> = new Map()
    v121: Map<string, number> = new Map()
    v122: Map<string, number | null> = new Map()
    v123: Map<string, Date | null> = new Map()
    v124: Map<string, Date | null> = new Map()
    v125: Map<string, string> = new Map()
    v126: Map<string, string | null> = new Map()
    v127: Map<string, HBufSubTestTs> = new Map()
    v128: Map<string, HBufSubTestTs | null> = new Map()

    descriptors(): Descriptor {
        return hbTestDesc
    }
}

const hbTestDesc: DataDescriptor = NewDataDescriptor(
    () => new HBufTestTs(),
    new Map<number, FieldDesc>([

        [1, {desc: NewInt8Descriptor(false), get: (o: HBufTestTs) => o.v1, set: (o: HBufTestTs, v) => o.v1 = v}],
        [2, {desc: NewInt8Descriptor(true), get: (o: HBufTestTs) => o.v2, set: (o: HBufTestTs, v) => o.v2 = v}],
        [3, {desc: NewInt16Descriptor(false), get: (o: HBufTestTs) => o.v3, set: (o: HBufTestTs, v) => o.v3 = v}],
        [4, {desc: NewInt16Descriptor(true), get: (o: HBufTestTs) => o.v4, set: (o: HBufTestTs, v) => o.v4 = v}],
        [5, {desc: NewInt32Descriptor(false), get: (o: HBufTestTs) => o.v5, set: (o: HBufTestTs, v) => o.v5 = v}],
        [6, {desc: NewInt32Descriptor(true), get: (o: HBufTestTs) => o.v6, set: (o: HBufTestTs, v) => o.v6 = v}],
        [7, {desc: NewInt64Descriptor(false), get: (o: HBufTestTs) => o.v7, set: (o: HBufTestTs, v) => o.v7 = v}],
        [8, {desc: NewInt64Descriptor(true), get: (o: HBufTestTs) => o.v8, set: (o: HBufTestTs, v) => o.v8 = v}],
        [9, {desc: NewUint8Descriptor(false), get: (o: HBufTestTs) => o.v9, set: (o: HBufTestTs, v) => o.v9 = v}],
        [10, {desc: NewUint8Descriptor(true), get: (o: HBufTestTs) => o.v10, set: (o: HBufTestTs, v) => o.v10 = v}],
        [11, {desc: NewUint16Descriptor(false), get: (o: HBufTestTs) => o.v11, set: (o: HBufTestTs, v) => o.v11 = v}],
        [12, {desc: NewUint16Descriptor(true), get: (o: HBufTestTs) => o.v12, set: (o: HBufTestTs, v) => o.v12 = v}],
        [13, {desc: NewUint32Descriptor(false), get: (o: HBufTestTs) => o.v13, set: (o: HBufTestTs, v) => o.v13 = v}],
        [14, {desc: NewUint32Descriptor(true), get: (o: HBufTestTs) => o.v14, set: (o: HBufTestTs, v) => o.v14 = v}],
        [15, {desc: NewUint64Descriptor(false), get: (o: HBufTestTs) => o.v15, set: (o: HBufTestTs, v) => o.v15 = v}],
        [16, {desc: NewUint64Descriptor(true), get: (o: HBufTestTs) => o.v16, set: (o: HBufTestTs, v) => o.v16 = v}],
        [17, {desc: NewBoolDescriptor(false), get: (o: HBufTestTs) => o.v17, set: (o: HBufTestTs, v) => o.v17 = v}],
        [18, {desc: NewBoolDescriptor(true), get: (o: HBufTestTs) => o.v18, set: (o: HBufTestTs, v) => o.v18 = v}],
        [19, {desc: NewStringDescriptor(false), get: (o: HBufTestTs) => o.v19, set: (o: HBufTestTs, v) => o.v19 = v}],
        [20, {desc: NewStringDescriptor(true), get: (o: HBufTestTs) => o.v20, set: (o: HBufTestTs, v) => o.v20 = v}],
        [21, {desc: NewBytesDescriptor(false), get: (o: HBufTestTs) => o.v21, set: (o: HBufTestTs, v) => o.v21 = v}],
        [22, {desc: NewBytesDescriptor(true), get: (o: HBufTestTs) => o.v22, set: (o: HBufTestTs, v) => o.v22 = v}],
        [23, {desc: NewFloatDescriptor(false), get: (o: HBufTestTs) => o.v23, set: (o: HBufTestTs, v) => o.v23 = v}],
        [24, {desc: NewFloatDescriptor(true), get: (o: HBufTestTs) => o.v24, set: (o: HBufTestTs, v) => o.v24 = v}],
        [25, {desc: NewDoubleDescriptor(false), get: (o: HBufTestTs) => o.v25, set: (o: HBufTestTs, v) => o.v25 = v}],
        [26, {desc: NewDoubleDescriptor(true), get: (o: HBufTestTs) => o.v26, set: (o: HBufTestTs, v) => o.v26 = v}],
        [27, {desc: NewTimeDescriptor(false), get: (o: HBufTestTs) => o.v27, set: (o: HBufTestTs, v) => o.v27 = v}],
        [28, {desc: NewTimeDescriptor(true), get: (o: HBufTestTs) => o.v28, set: (o: HBufTestTs, v) => o.v28 = v}],
        [29, {desc: NewDecimalDescriptor(false), get: (o: HBufTestTs) => o.v29, set: (o: HBufTestTs, v) => o.v29 = v}],
        [30, {desc: NewDecimalDescriptor(true), get: (o: HBufTestTs) => o.v30, set: (o: HBufTestTs, v) => o.v30 = v}],
        [31, {desc: NewRefDescriptor(hbSubDesc, false), get: (o: HBufTestTs) => o.v31, set: (o: HBufTestTs, v) => o.v31 = v}],
        [32, {desc: NewRefDescriptor(hbSubDesc, true), get: (o: HBufTestTs) => o.v32, set: (o: HBufTestTs, v) => o.v32 = v}],
        [33, {desc: NewListDescriptor(NewInt8Descriptor(false), false), get: (o: HBufTestTs) => o.v33, set: (o: HBufTestTs, v) => o.v33 = v}],
        [34, {desc: NewListDescriptor(NewInt8Descriptor(true), false), get: (o: HBufTestTs) => o.v34, set: (o: HBufTestTs, v) => o.v34 = v}],
        [35, {desc: NewListDescriptor(NewInt16Descriptor(false), false), get: (o: HBufTestTs) => o.v35, set: (o: HBufTestTs, v) => o.v35 = v}],
        [36, {desc: NewListDescriptor(NewInt16Descriptor(true), false), get: (o: HBufTestTs) => o.v36, set: (o: HBufTestTs, v) => o.v36 = v}],
        [37, {desc: NewListDescriptor(NewInt32Descriptor(false), false), get: (o: HBufTestTs) => o.v37, set: (o: HBufTestTs, v) => o.v37 = v}],
        [38, {desc: NewListDescriptor(NewInt32Descriptor(true), false), get: (o: HBufTestTs) => o.v38, set: (o: HBufTestTs, v) => o.v38 = v}],
        [39, {desc: NewListDescriptor(NewInt64Descriptor(false), false), get: (o: HBufTestTs) => o.v39, set: (o: HBufTestTs, v) => o.v39 = v}],
        [40, {desc: NewListDescriptor(NewInt64Descriptor(true), false), get: (o: HBufTestTs) => o.v40, set: (o: HBufTestTs, v) => o.v40 = v}],
        [41, {desc: NewListDescriptor(NewUint8Descriptor(false), false), get: (o: HBufTestTs) => o.v41, set: (o: HBufTestTs, v) => o.v41 = v}],
        [42, {desc: NewListDescriptor(NewUint8Descriptor(true), false), get: (o: HBufTestTs) => o.v42, set: (o: HBufTestTs, v) => o.v42 = v}],
        [43, {desc: NewListDescriptor(NewUint16Descriptor(false), false), get: (o: HBufTestTs) => o.v43, set: (o: HBufTestTs, v) => o.v43 = v}],
        [44, {desc: NewListDescriptor(NewUint16Descriptor(true), false), get: (o: HBufTestTs) => o.v44, set: (o: HBufTestTs, v) => o.v44 = v}],
        [45, {desc: NewListDescriptor(NewUint32Descriptor(false), false), get: (o: HBufTestTs) => o.v45, set: (o: HBufTestTs, v) => o.v45 = v}],
        [46, {desc: NewListDescriptor(NewUint32Descriptor(true), false), get: (o: HBufTestTs) => o.v46, set: (o: HBufTestTs, v) => o.v46 = v}],
        [47, {desc: NewListDescriptor(NewUint64Descriptor(false), false), get: (o: HBufTestTs) => o.v47, set: (o: HBufTestTs, v) => o.v47 = v}],
        [48, {desc: NewListDescriptor(NewUint64Descriptor(true), false), get: (o: HBufTestTs) => o.v48, set: (o: HBufTestTs, v) => o.v48 = v}],
        [49, {desc: NewListDescriptor(NewBoolDescriptor(false), false), get: (o: HBufTestTs) => o.v49, set: (o: HBufTestTs, v) => o.v49 = v}],
        [50, {desc: NewListDescriptor(NewBoolDescriptor(true), false), get: (o: HBufTestTs) => o.v50, set: (o: HBufTestTs, v) => o.v50 = v}],
        [51, {desc: NewListDescriptor(NewStringDescriptor(false), false), get: (o: HBufTestTs) => o.v51, set: (o: HBufTestTs, v) => o.v51 = v}],
        [52, {desc: NewListDescriptor(NewStringDescriptor(true), false), get: (o: HBufTestTs) => o.v52, set: (o: HBufTestTs, v) => o.v52 = v}],
        [53, {desc: NewListDescriptor(NewBytesDescriptor(false), false), get: (o: HBufTestTs) => o.v53, set: (o: HBufTestTs, v) => o.v53 = v}],
        [54, {desc: NewListDescriptor(NewBytesDescriptor(true), false), get: (o: HBufTestTs) => o.v54, set: (o: HBufTestTs, v) => o.v54 = v}],
        [55, {desc: NewListDescriptor(NewFloatDescriptor(false), false), get: (o: HBufTestTs) => o.v55, set: (o: HBufTestTs, v) => o.v55 = v}],
        [56, {desc: NewListDescriptor(NewFloatDescriptor(true), false), get: (o: HBufTestTs) => o.v56, set: (o: HBufTestTs, v) => o.v56 = v}],
        [57, {desc: NewListDescriptor(NewDoubleDescriptor(false), false), get: (o: HBufTestTs) => o.v57, set: (o: HBufTestTs, v) => o.v57 = v}],
        [58, {desc: NewListDescriptor(NewDoubleDescriptor(true), false), get: (o: HBufTestTs) => o.v58, set: (o: HBufTestTs, v) => o.v58 = v}],
        [59, {desc: NewListDescriptor(NewTimeDescriptor(false), false), get: (o: HBufTestTs) => o.v59, set: (o: HBufTestTs, v) => o.v59 = v}],
        [60, {desc: NewListDescriptor(NewTimeDescriptor(true), false), get: (o: HBufTestTs) => o.v60, set: (o: HBufTestTs, v) => o.v60 = v}],
        [61, {desc: NewListDescriptor(NewDecimalDescriptor(false), false), get: (o: HBufTestTs) => o.v61, set: (o: HBufTestTs, v) => o.v61 = v}],
        [62, {desc: NewListDescriptor(NewDecimalDescriptor(true), false), get: (o: HBufTestTs) => o.v62, set: (o: HBufTestTs, v) => o.v62 = v}],
        [63, {desc: NewListDescriptor(NewRefDescriptor(hbSubDesc, false), false), get: (o: HBufTestTs) => o.v63, set: (o: HBufTestTs, v) => o.v63 = v}],
        [64, {desc: NewListDescriptor(NewRefDescriptor(hbSubDesc, true), false), get: (o: HBufTestTs) => o.v64, set: (o: HBufTestTs, v) => o.v64 = v}],
        [65, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt8Descriptor(false), false), get: (o: HBufTestTs) => o.v65, set: (o: HBufTestTs, v) => o.v65 = v}],
        [66, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt8Descriptor(true), false), get: (o: HBufTestTs) => o.v66, set: (o: HBufTestTs, v) => o.v66 = v}],
        [67, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt16Descriptor(false), false), get: (o: HBufTestTs) => o.v67, set: (o: HBufTestTs, v) => o.v67 = v}],
        [68, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt16Descriptor(true), false), get: (o: HBufTestTs) => o.v68, set: (o: HBufTestTs, v) => o.v68 = v}],
        [69, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt32Descriptor(false), false), get: (o: HBufTestTs) => o.v69, set: (o: HBufTestTs, v) => o.v69 = v}],
        [70, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt32Descriptor(true), false), get: (o: HBufTestTs) => o.v70, set: (o: HBufTestTs, v) => o.v70 = v}],
        [71, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt64Descriptor(false), false), get: (o: HBufTestTs) => o.v71, set: (o: HBufTestTs, v) => o.v71 = v}],
        [72, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewInt64Descriptor(true), false), get: (o: HBufTestTs) => o.v72, set: (o: HBufTestTs, v) => o.v72 = v}],
        [73, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewUint8Descriptor(false), false), get: (o: HBufTestTs) => o.v73, set: (o: HBufTestTs, v) => o.v73 = v}],
        [74, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewUint8Descriptor(true), false), get: (o: HBufTestTs) => o.v74, set: (o: HBufTestTs, v) => o.v74 = v}],
        [75, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewUint16Descriptor(false), false), get: (o: HBufTestTs) => o.v75, set: (o: HBufTestTs, v) => o.v75 = v}],
        [76, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewUint16Descriptor(true), false), get: (o: HBufTestTs) => o.v76, set: (o: HBufTestTs, v) => o.v76 = v}],
        [77, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewUint32Descriptor(false), false), get: (o: HBufTestTs) => o.v77, set: (o: HBufTestTs, v) => o.v77 = v}],
        [78, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewUint32Descriptor(true), false), get: (o: HBufTestTs) => o.v78, set: (o: HBufTestTs, v) => o.v78 = v}],
        [79, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewUint64Descriptor(false), false), get: (o: HBufTestTs) => o.v79, set: (o: HBufTestTs, v) => o.v79 = v}],
        [80, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewUint64Descriptor(true), false), get: (o: HBufTestTs) => o.v80, set: (o: HBufTestTs, v) => o.v80 = v}],
        [81, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewBoolDescriptor(false), false), get: (o: HBufTestTs) => o.v81, set: (o: HBufTestTs, v) => o.v81 = v}],
        [82, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewBoolDescriptor(true), false), get: (o: HBufTestTs) => o.v82, set: (o: HBufTestTs, v) => o.v82 = v}],
        [83, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewStringDescriptor(false), false), get: (o: HBufTestTs) => o.v83, set: (o: HBufTestTs, v) => o.v83 = v}],
        [84, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewStringDescriptor(true), false), get: (o: HBufTestTs) => o.v84, set: (o: HBufTestTs, v) => o.v84 = v}],
        [85, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewBytesDescriptor(false), false), get: (o: HBufTestTs) => o.v85, set: (o: HBufTestTs, v) => o.v85 = v}],
        [86, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewBytesDescriptor(true), false), get: (o: HBufTestTs) => o.v86, set: (o: HBufTestTs, v) => o.v86 = v}],
        [87, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewFloatDescriptor(false), false), get: (o: HBufTestTs) => o.v87, set: (o: HBufTestTs, v) => o.v87 = v}],
        [88, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewFloatDescriptor(true), false), get: (o: HBufTestTs) => o.v88, set: (o: HBufTestTs, v) => o.v88 = v}],
        [89, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewDoubleDescriptor(false), false), get: (o: HBufTestTs) => o.v89, set: (o: HBufTestTs, v) => o.v89 = v}],
        [90, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewDoubleDescriptor(true), false), get: (o: HBufTestTs) => o.v90, set: (o: HBufTestTs, v) => o.v90 = v}],
        [91, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewTimeDescriptor(false), false), get: (o: HBufTestTs) => o.v91, set: (o: HBufTestTs, v) => o.v91 = v}],
        [92, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewTimeDescriptor(true), false), get: (o: HBufTestTs) => o.v92, set: (o: HBufTestTs, v) => o.v92 = v}],
        [93, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewDecimalDescriptor(false), false), get: (o: HBufTestTs) => o.v93, set: (o: HBufTestTs, v) => o.v93 = v}],
        [94, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewDecimalDescriptor(true), false), get: (o: HBufTestTs) => o.v94, set: (o: HBufTestTs, v) => o.v94 = v}],
        [95, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewRefDescriptor(hbSubDesc, false), false), get: (o: HBufTestTs) => o.v95, set: (o: HBufTestTs, v) => o.v95 = v}],
        [96, {desc: NewMapDescriptor(NewInt64Descriptor(false), NewRefDescriptor(hbSubDesc, true), false), get: (o: HBufTestTs) => o.v96, set: (o: HBufTestTs, v) => o.v96 = v}],
        [97, {desc: NewMapDescriptor(NewStringDescriptor(false), NewInt8Descriptor(false), false), get: (o: HBufTestTs) => o.v97, set: (o: HBufTestTs, v) => o.v97 = v}],
        [98, {desc: NewMapDescriptor(NewStringDescriptor(false), NewInt8Descriptor(true), false), get: (o: HBufTestTs) => o.v98, set: (o: HBufTestTs, v) => o.v98 = v}],
        [99, {desc: NewMapDescriptor(NewStringDescriptor(false), NewInt16Descriptor(false), false), get: (o: HBufTestTs) => o.v99, set: (o: HBufTestTs, v) => o.v99 = v}],
        [100, {desc: NewMapDescriptor(NewStringDescriptor(false), NewInt16Descriptor(true), false), get: (o: HBufTestTs) => o.v100, set: (o: HBufTestTs, v) => o.v100 = v}],
        [101, {desc: NewMapDescriptor(NewStringDescriptor(false), NewInt32Descriptor(false), false), get: (o: HBufTestTs) => o.v101, set: (o: HBufTestTs, v) => o.v101 = v}],
        [102, {desc: NewMapDescriptor(NewStringDescriptor(false), NewInt32Descriptor(true), false), get: (o: HBufTestTs) => o.v102, set: (o: HBufTestTs, v) => o.v102 = v}],
        [103, {desc: NewMapDescriptor(NewStringDescriptor(false), NewInt64Descriptor(false), false), get: (o: HBufTestTs) => o.v103, set: (o: HBufTestTs, v) => o.v103 = v}],
        [104, {desc: NewMapDescriptor(NewStringDescriptor(false), NewInt64Descriptor(true), false), get: (o: HBufTestTs) => o.v104, set: (o: HBufTestTs, v) => o.v104 = v}],
        [105, {desc: NewMapDescriptor(NewStringDescriptor(false), NewUint8Descriptor(false), false), get: (o: HBufTestTs) => o.v105, set: (o: HBufTestTs, v) => o.v105 = v}],
        [106, {desc: NewMapDescriptor(NewStringDescriptor(false), NewUint8Descriptor(true), false), get: (o: HBufTestTs) => o.v106, set: (o: HBufTestTs, v) => o.v106 = v}],
        [107, {desc: NewMapDescriptor(NewStringDescriptor(false), NewUint16Descriptor(false), false), get: (o: HBufTestTs) => o.v107, set: (o: HBufTestTs, v) => o.v107 = v}],
        [108, {desc: NewMapDescriptor(NewStringDescriptor(false), NewUint16Descriptor(true), false), get: (o: HBufTestTs) => o.v108, set: (o: HBufTestTs, v) => o.v108 = v}],
        [109, {desc: NewMapDescriptor(NewStringDescriptor(false), NewUint32Descriptor(false), false), get: (o: HBufTestTs) => o.v109, set: (o: HBufTestTs, v) => o.v109 = v}],
        [110, {desc: NewMapDescriptor(NewStringDescriptor(false), NewUint32Descriptor(true), false), get: (o: HBufTestTs) => o.v110, set: (o: HBufTestTs, v) => o.v110 = v}],
        [111, {desc: NewMapDescriptor(NewStringDescriptor(false), NewUint64Descriptor(false), false), get: (o: HBufTestTs) => o.v111, set: (o: HBufTestTs, v) => o.v111 = v}],
        [112, {desc: NewMapDescriptor(NewStringDescriptor(false), NewUint64Descriptor(true), false), get: (o: HBufTestTs) => o.v112, set: (o: HBufTestTs, v) => o.v112 = v}],
        [113, {desc: NewMapDescriptor(NewStringDescriptor(false), NewBoolDescriptor(false), false), get: (o: HBufTestTs) => o.v113, set: (o: HBufTestTs, v) => o.v113 = v}],
        [114, {desc: NewMapDescriptor(NewStringDescriptor(false), NewBoolDescriptor(true), false), get: (o: HBufTestTs) => o.v114, set: (o: HBufTestTs, v) => o.v114 = v}],
        [115, {desc: NewMapDescriptor(NewStringDescriptor(false), NewStringDescriptor(false), false), get: (o: HBufTestTs) => o.v115, set: (o: HBufTestTs, v) => o.v115 = v}],
        [116, {desc: NewMapDescriptor(NewStringDescriptor(false), NewStringDescriptor(true), false), get: (o: HBufTestTs) => o.v116, set: (o: HBufTestTs, v) => o.v116 = v}],
        [117, {desc: NewMapDescriptor(NewStringDescriptor(false), NewBytesDescriptor(false), false), get: (o: HBufTestTs) => o.v117, set: (o: HBufTestTs, v) => o.v117 = v}],
        [118, {desc: NewMapDescriptor(NewStringDescriptor(false), NewBytesDescriptor(true), false), get: (o: HBufTestTs) => o.v118, set: (o: HBufTestTs, v) => o.v118 = v}],
        [119, {desc: NewMapDescriptor(NewStringDescriptor(false), NewFloatDescriptor(false), false), get: (o: HBufTestTs) => o.v119, set: (o: HBufTestTs, v) => o.v119 = v}],
        [120, {desc: NewMapDescriptor(NewStringDescriptor(false), NewFloatDescriptor(true), false), get: (o: HBufTestTs) => o.v120, set: (o: HBufTestTs, v) => o.v120 = v}],
        [121, {desc: NewMapDescriptor(NewStringDescriptor(false), NewDoubleDescriptor(false), false), get: (o: HBufTestTs) => o.v121, set: (o: HBufTestTs, v) => o.v121 = v}],
        [122, {desc: NewMapDescriptor(NewStringDescriptor(false), NewDoubleDescriptor(true), false), get: (o: HBufTestTs) => o.v122, set: (o: HBufTestTs, v) => o.v122 = v}],
        [123, {desc: NewMapDescriptor(NewStringDescriptor(false), NewTimeDescriptor(false), false), get: (o: HBufTestTs) => o.v123, set: (o: HBufTestTs, v) => o.v123 = v}],
        [124, {desc: NewMapDescriptor(NewStringDescriptor(false), NewTimeDescriptor(true), false), get: (o: HBufTestTs) => o.v124, set: (o: HBufTestTs, v) => o.v124 = v}],
        [125, {desc: NewMapDescriptor(NewStringDescriptor(false), NewDecimalDescriptor(false), false), get: (o: HBufTestTs) => o.v125, set: (o: HBufTestTs, v) => o.v125 = v}],
        [126, {desc: NewMapDescriptor(NewStringDescriptor(false), NewDecimalDescriptor(true), false), get: (o: HBufTestTs) => o.v126, set: (o: HBufTestTs, v) => o.v126 = v}],
        [127, {desc: NewMapDescriptor(NewStringDescriptor(false), NewRefDescriptor(hbSubDesc, false), false), get: (o: HBufTestTs) => o.v127, set: (o: HBufTestTs, v) => o.v127 = v}],
        [128, {desc: NewMapDescriptor(NewStringDescriptor(false), NewRefDescriptor(hbSubDesc, true), false), get: (o: HBufTestTs) => o.v128, set: (o: HBufTestTs, v) => o.v128 = v}],
    ]),
)

// ===================== 填值（翻译自 Go data_test.go 的 src，语义一致） =====================

const longStr = "Hello,世界! HBuf/proto 对比测试 ".repeat(8)
const fullBytes = Uint8Array.from({length: 256}, (_, i) => i)
const testTime = new Date(1234567890 * 1000 + 987) // Go: Unix(1234567890, 987654321) → TS 毫秒截断
const maxTime = new Date(9999, 11, 31, 23, 59, 59, 999) // Go: Date(9999,...,999999999, Local)
const bigDec = "123456789012345678901234567890.123456789"
const smallDec = "-0.0000000000000000000000000001"
const fMax32 = Math.fround(3.4028234663852886e38)

function sub(a1: bigint, a2: bigint | null): HBufSubTestTs {
    const s = new HBufSubTestTs()
    s.a1 = a1
    s.a2 = a2
    return s
}

function buildFull(): HBufTestTs {
    const v = new HBufTestTs()

    v.v1 = 127
    v.v2 = -128
    v.v3 = 32767
    v.v4 = -32768
    v.v5 = 2147483647
    v.v6 = -2147483648
    v.v7 = 9223372036854775807n
    v.v8 = -9223372036854775808n
    v.v9 = 255
    v.v10 = 255
    v.v11 = 65535
    v.v12 = 65535
    v.v13 = 4294967295
    v.v14 = 4294967295
    v.v15 = 18446744073709551615n
    v.v16 = 18446744073709551615n
    v.v17 = true
    v.v18 = false // 非 nil 的 false 指针也写入
    v.v19 = longStr
    v.v20 = "" // 空串指针
    v.v21 = fullBytes
    v.v22 = new Uint8Array(0) // 空 bytes 指针
    v.v23 = fMax32
    v.v24 = Math.fround(-3.4028234663852886e38)
    v.v25 = 1.7976931348623157e308
    v.v26 = -1.7976931348623157e308
    v.v27 = maxTime
    v.v28 = testTime
    v.v29 = bigDec
    v.v30 = smallDec
    v.v31 = sub(2147483647n, -2147483648n)
    v.v32 = sub(-2147483648n, 2147483647n)
    v.v33 = [-128, -1, 0, 1, 127]
    v.v34 = [-128, 127]
    v.v35 = [-32768, -1, 0, 1, 32767]
    v.v36 = [-32768, 32767]
    v.v37 = [-2147483648, -1, 0, 1, 2147483647]
    v.v38 = [-2147483648, 2147483647]
    v.v39 = [-9223372036854775808n, -1n, 0n, 1n, 9223372036854775807n]
    v.v40 = [-9223372036854775808n, 9223372036854775807n]
    v.v41 = [0, 1, 127, 128, 255]
    v.v42 = [1, 255]
    v.v43 = [0, 1, 255, 256, 65535]
    v.v44 = [1, 65535]
    v.v45 = [0, 1, 4294967295]
    v.v46 = [1, 4294967295]
    v.v47 = [0n, 1n, 18446744073709551615n]
    v.v48 = [1n, 18446744073709551615n]
    v.v49 = [true, false]
    v.v50 = [true, false]
    v.v51 = ["", "hello world", longStr]
    v.v52 = ["hello world", longStr]
    v.v53 = [new Uint8Array([0]), fullBytes]
    v.v54 = [new Uint8Array([0]), fullBytes]
    v.v55 = [Math.fround(-3.4028234663852886e38), -1.5, 0, 1.5, fMax32]
    v.v56 = [Math.fround(-3.4028234663852886e38), fMax32]
    v.v57 = [-1.7976931348623157e308, -1.5, 0, 1.5, 1.7976931348623157e308]
    v.v58 = [-1.7976931348623157e308, 1.7976931348623157e308]
    v.v59 = [null, testTime, maxTime] // time.Time{} 零值 → null
    v.v60 = [testTime, maxTime]
    v.v61 = ["0", smallDec, bigDec] // decimal.Zero → "0"
    v.v62 = [smallDec, bigDec]
    v.v63 = [sub(0n, null), sub(2147483647n, -2147483648n), sub(-2147483648n, 2147483647n)]
    v.v64 = [sub(0n, null), sub(2147483647n, -2147483648n)]
    // int64 key 的 map
    v.v65 = new Map([[-9223372036854775808n, -128], [-1n, -1], [0n, 0], [1n, 127], [9223372036854775807n, -128]])
    v.v66 = new Map([[-9223372036854775808n, -128], [1n, 127]])
    v.v67 = new Map([[-9223372036854775808n, -32768], [-1n, -1], [0n, 0], [1n, 32767], [9223372036854775807n, -32768]])
    v.v68 = new Map([[-9223372036854775808n, -32768], [1n, 32767]])
    v.v69 = new Map([[-9223372036854775808n, -2147483648], [-1n, -1], [0n, 0], [1n, 2147483647], [9223372036854775807n, -2147483648]])
    v.v70 = new Map([[-9223372036854775808n, -2147483648], [1n, 2147483647]])
    v.v71 = new Map([[-9223372036854775808n, -9223372036854775808n], [-1n, -1n], [0n, 0n], [1n, 9223372036854775807n], [9223372036854775807n, 9223372036854775807n]])
    v.v72 = new Map([[-9223372036854775808n, -9223372036854775808n], [1n, 9223372036854775807n]])
    v.v73 = new Map([[-9223372036854775808n, 0], [-1n, 1], [0n, 0], [1n, 255], [9223372036854775807n, 255]])
    v.v74 = new Map([[-9223372036854775808n, 1], [1n, 255]])
    v.v75 = new Map([[-9223372036854775808n, 0], [-1n, 1], [0n, 0], [1n, 65535], [9223372036854775807n, 65535]])
    v.v76 = new Map([[-9223372036854775808n, 1], [1n, 65535]])
    v.v77 = new Map([[-9223372036854775808n, 0], [-1n, 1], [0n, 0], [1n, 4294967295], [9223372036854775807n, 4294967295]])
    v.v78 = new Map([[-9223372036854775808n, 1], [1n, 4294967295]])
    v.v79 = new Map([[-9223372036854775808n, 0n], [-1n, 1n], [0n, 0n], [1n, 18446744073709551615n], [9223372036854775807n, 18446744073709551615n]])
    v.v80 = new Map([[-9223372036854775808n, 1n], [1n, 18446744073709551615n]])
    v.v81 = new Map([[-9223372036854775808n, false], [-1n, false], [0n, false], [1n, true], [9223372036854775807n, true]])
    v.v82 = new Map([[-9223372036854775808n, false], [0n, false], [1n, true], [9223372036854775807n, true]])
    v.v83 = new Map([[-9223372036854775808n, ""], [-1n, ""], [0n, ""], [1n, "hello world"], [9223372036854775807n, longStr]])
    v.v84 = new Map([[-9223372036854775808n, "hello world"], [0n, longStr], [1n, "hello world"]])
    v.v85 = new Map([[-9223372036854775808n, new Uint8Array([0])], [-1n, new Uint8Array([0])], [0n, new Uint8Array([0])], [1n, fullBytes], [9223372036854775807n, fullBytes]])
    v.v86 = new Map([[-9223372036854775808n, new Uint8Array([0])], [1n, fullBytes]])
    v.v87 = new Map([[-9223372036854775808n, Math.fround(-3.4028234663852886e38)], [-1n, -1.5], [0n, 0], [1n, 1.5], [9223372036854775807n, fMax32]])
    v.v88 = new Map([[-9223372036854775808n, Math.fround(-3.4028234663852886e38)], [1n, fMax32]])
    v.v89 = new Map([[-9223372036854775808n, -1.7976931348623157e308], [-1n, -1.5], [0n, 0], [1n, 1.5], [9223372036854775807n, 1.7976931348623157e308]])
    v.v90 = new Map([[-9223372036854775808n, -1.7976931348623157e308], [1n, 1.7976931348623157e308]])
    v.v91 = new Map([[-9223372036854775808n, null], [-1n, null], [0n, testTime], [1n, maxTime], [9223372036854775807n, maxTime]])
    v.v92 = new Map([[-9223372036854775808n, testTime], [0n, maxTime]])
    v.v93 = new Map([[-9223372036854775808n, "0"], [-1n, "0"], [0n, smallDec], [1n, bigDec], [9223372036854775807n, bigDec]])
    v.v94 = new Map([[-9223372036854775808n, smallDec], [0n, bigDec]])
    v.v95 = new Map([[-9223372036854775808n, sub(0n, null)], [-1n, sub(0n, null)], [0n, sub(2147483647n, -2147483648n)], [1n, sub(-2147483648n, 2147483647n)], [9223372036854775807n, sub(0n, null)]])
    v.v96 = new Map([[-9223372036854775808n, sub(0n, null)], [0n, sub(2147483647n, -2147483648n)], [1n, sub(-2147483648n, 2147483647n)]])
    // string key 的 map
    v.v97 = new Map([["", -128], ["a", 0], [longStr, 127]])
    v.v98 = new Map([["", -128], [longStr, 127]])
    v.v99 = new Map([["", -32768], ["a", 0], [longStr, 32767]])
    v.v100 = new Map([["", -32768], [longStr, 32767]])
    v.v101 = new Map([["", -2147483648], ["a", 0], [longStr, 2147483647]])
    v.v102 = new Map([["", -2147483648], [longStr, 2147483647]])
    v.v103 = new Map([["", -9223372036854775808n], ["a", 0n], [longStr, 9223372036854775807n]])
    v.v104 = new Map([["", -9223372036854775808n], [longStr, 9223372036854775807n]])
    v.v105 = new Map([["", 0], ["a", 0], [longStr, 255]])
    v.v106 = new Map([["", 1], [longStr, 255]])
    v.v107 = new Map([["", 0], ["a", 0], [longStr, 65535]])
    v.v108 = new Map([["", 1], [longStr, 65535]])
    v.v109 = new Map([["", 0], ["a", 0], [longStr, 4294967295]])
    v.v110 = new Map([["", 1], [longStr, 4294967295]])
    v.v111 = new Map([["", 0n], ["a", 0n], [longStr, 18446744073709551615n]])
    v.v112 = new Map([["", 1n], [longStr, 18446744073709551615n]])
    v.v113 = new Map([["", false], ["a", false], [longStr, true]])
    v.v114 = new Map([["", false], ["a", false], [longStr, true]])
    v.v115 = new Map([["", ""], ["a", "hello world"], [longStr, longStr]])
    v.v116 = new Map([["a", "hello world"], [longStr, longStr]])
    v.v117 = new Map([["", new Uint8Array([0])], ["a", fullBytes], [longStr, new Uint8Array([0])]])
    v.v118 = new Map([["a", fullBytes], [longStr, new Uint8Array([0])]])
    v.v119 = new Map([["", Math.fround(-3.4028234663852886e38)], ["a", 0], [longStr, fMax32]])
    v.v120 = new Map([["", Math.fround(-3.4028234663852886e38)], [longStr, fMax32]])
    v.v121 = new Map([["", -1.7976931348623157e308], ["a", 0], [longStr, 1.7976931348623157e308]])
    v.v122 = new Map([["", -1.7976931348623157e308], [longStr, 1.7976931348623157e308]])
    v.v123 = new Map([["", null], ["a", testTime], [longStr, maxTime]])
    v.v124 = new Map([["a", testTime], [longStr, maxTime]])
    v.v125 = new Map([["", "0"], ["a", smallDec], [longStr, bigDec]])
    v.v126 = new Map([["a", smallDec], [longStr, bigDec]])
    v.v127 = new Map([["", sub(0n, null)], ["a", sub(2147483647n, -2147483648n)], [longStr, sub(-2147483648n, 2147483647n)]])
    v.v128 = new Map([["", sub(0n, null)], ["a", sub(2147483647n, -2147483648n)], [longStr, sub(-2147483648n, 2147483647n)]])
    return v
}

// nomap: 清空 map 字段（时间字段 Go 侧已替换为整毫秒，TS Date 值不变）
function buildNoMap(): HBufTestTs {
    const v = buildFull()
    for (let i = 65; i <= 128; i++) {
        ;(v as any)["v" + i] = new Map()
    }
    return v
}

function bytesToHex(b: Uint8Array): string {
    return Array.from(b, (x) => x.toString(16).padStart(2, "0")).join("")
}

function hexToBytes(hex: string): Uint8Array {
    const ret = new Uint8Array(hex.length / 2)
    for (let i = 0; i < hex.length; i += 2) {
        ret[i / 2] = parseInt(hex.substring(i, i + 2), 16)
    }
    return ret
}

function deepEq(a: any, b: any, path = "$"): void {
    if (a === b) return
    if (a == null || b == null) {
        assert.strictEqual(a, b, path)
        return
    }
    if (typeof a !== typeof b) {
        assert.strictEqual(typeof a, typeof b, path + " typeof")
        return
    }
    if (a instanceof Date && b instanceof Date) {
        assert.strictEqual(a.getTime(), b.getTime(), path)
        return
    }
    if (a instanceof Uint8Array && b instanceof Uint8Array) {
        assert.strictEqual(a.length, b.length, path + " len")
        for (let i = 0; i < a.length; i++) {
            assert.strictEqual(a[i], b[i], path + "[" + i + "]")
        }
        return
    }
    if (a instanceof Map && b instanceof Map) {
        assert.strictEqual(a.size, b.size, path + " size")
        for (const [k, v] of a) {
            assert.ok(b.has(k), path + " key " + k)
            deepEq(v, b.get(k), path + "[" + k + "]")
        }
        return
    }
    if (Array.isArray(a) && Array.isArray(b)) {
        assert.strictEqual(a.length, b.length, path + " len")
        for (let i = 0; i < a.length; i++) {
            deepEq(a[i], b[i], path + "[" + i + "]")
        }
        return
    }
    if (typeof a === "object") {
        for (const k of Object.keys(a)) {
            deepEq(a[k], b[k], path + "." + k)
        }
        return
    }
    assert.strictEqual(a, b, path)
}

// ===================== 测试 =====================

test("HBufTest 128 字段: TS 编码与 Go 字节级一致（无 map 场景）", () => {
    const nomapHex = fs.readFileSync(new URL("./golden_nomap.hex", import.meta.url), "utf-8").trim()
    const buf = Marshal(buildNoMap())
    assert.strictEqual(bytesToHex(buf), nomapHex)
})

test("HBufTest 128 字段: TS 解码 Go full 数据，128 字段语义一致", () => {
    const fullHex = fs.readFileSync(new URL("./golden_full.hex", import.meta.url), "utf-8").trim()
    const dst = new HBufTestTs()
    Unmarshal(hexToBytes(fullHex), dst)
    deepEq(dst, buildFull())
})

test("HBufTest 128 字段: TS 编码长度与 Go 一致", () => {
    assert.strictEqual(Marshal(buildFull()).length, 16942)
    assert.strictEqual(Marshal(buildNoMap()).length, 2599)
})

test("HBufTest 128 字段: TS round-trip", () => {
    const src = buildFull()
    const dst = new HBufTestTs()
    Unmarshal(Marshal(src), dst)
    deepEq(dst, src)
})
