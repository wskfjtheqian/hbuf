# HBuf TS 代码生成器规范

本文件是 `.hbuf` IDL → TypeScript 数据类**代码生成器**的完整实现规范。按本文件实现的生成器，其产出代码可直接对接 `src/hbuf` 运行时，与 Go 版 `pkg/hbuf` **字节级兼容**（已通过 25 个对比测试验证，见 `test/`）。

## 1. 目标与验收标准

- **输入**：`.hbuf` IDL 文件（语法见第 2 节）。
- **输出**：`.ts` 文件，每个 `data` 生成一个类，实现 `HData` 接口。
- **验收**：
  1. `npx tsc --noEmit` 通过；
  2. 生成代码可被 `Marshal` / `Unmarshal` 正常编解码；
  3. 与 Go 版编码**字节级一致**（用第 9 节的 golden 方法验证）。

## 2. IDL 语法参考

```
package go = "hbuf"

data HBufSubTest {
    int64 A1 = 1
    int64? A2 = 2          // ? 表示可空（Go 指针 / TS null）
}

data HBufTest {
    int8 V1 = 1
    string? V20 = 20
    int8[] v33 = 33        // 列表
    int8?[] v34 = 34       // 元素可空列表
    int8<int64> v65 = 65   // map：VALUE<KEY>
    int8?<int64> v66 = 66  // 值可空 map
    HBufSubTest V31 = 31   // 引用其他 data
    HBufSubTest? V32 = 32
}

data HBufSub1Test:HBufBTest = 1, HBufSubTest = 2, HBufTest = 3 {
    int64 V1 = 1
    int64? V2 = 2
}
```

- 基础类型：`int8/16/32/64`、`uint8/16/32/64`、`bool`、`string`、`bytes`、`float`、`double`、`date`、`decimal`。
- 复合类型：`T[]`（列表）、`V<K>`（map，K 仅 `int64` 或 `string`）、data 引用。
- 可空 `T?` 可修饰标量、列表元素（`T?[]`）、map 值（`T?<K>`）。
- `extends`：`data Name:Parent = extId, ...`，每个父类型带一个扩展 ID（`extId`），父类型以**内嵌字段**形式存在于子类型中。
- 字段 ID：1 ~ 65535 正整数。

## 3. 运行时 API 契约（生成的代码如何对接）

生成代码只依赖以下导出（均从包入口 `src/index.ts` 导出）：

```ts
export interface HData {
    descriptors(): Descriptor
}

export interface Descriptor {
    zero(): any                                          // 元素缺失时的零值
    isEmpty(v: any): boolean                             // 字段空判定
    encode(buf: Writer, v: any, id?: number): void       // id 存在=字段编码，否则=元素编码
    decode(r: Reader, typ: Type, valRead: boolean, valueLen: number, target?: any): any
}

export type FieldDesc = {
    desc: Descriptor
    get: (o: any) => any      // 替代 Go unsafe.Offsetof 的读闭包
    set: (o: any, v: any) => void
}

export function Marshal(data: HData): Uint8Array
export function Unmarshal(buf: Uint8Array, data: HData): void   // 就地填充

// 构造器
NewDataDescriptor(factory: () => any, fields: Map<number, FieldDesc>,
                  extends_?: Map<number, FieldDesc>, isPtr?: boolean): DataDescriptor
NewRefDescriptor(desc: Descriptor, isPtr = false): RefDescriptor
NewSyncDescriptor(f: () => Descriptor): SyncDescriptor          // 递归/延迟构建
NewListDescriptor(desc: Descriptor, isPtr = false): ListDescriptor
NewMapDescriptor(keyDesc: Descriptor, valueDesc: Descriptor, isPtr = false): MapDescriptor
NewInt8Descriptor(isPtr = false)    NewInt16Descriptor(isPtr = false)
NewInt32Descriptor(isPtr = false)   NewInt64Descriptor(isPtr = false)
NewUint8Descriptor(isPtr = false)   NewUint16Descriptor(isPtr = false)
NewUint32Descriptor(isPtr = false)  NewUint64Descriptor(isPtr = false)
NewFloatDescriptor(isPtr = false)   NewDoubleDescriptor(isPtr = false)
NewBoolDescriptor(isPtr = false)    NewStringDescriptor(isPtr = false)
NewBytesDescriptor(isPtr = false)   NewTimeDescriptor(isPtr = false)
NewDecimalDescriptor(isPtr = false)
```

## 4. 类型映射表

| IDL 类型 | TS 类型 | Descriptor 工厂 | 说明 |
|---|---|---|---|
| `int8` / `int16` / `int32` | `number` | `NewInt8/16/32Descriptor(isPtr)` | 溢出按位截断（与 Go 一致） |
| `int64` | `bigint` | `NewInt64Descriptor(isPtr)` | 超出 2^53 保真 |
| `uint8` / `uint16` / `uint32` | `number` | `NewUint8/16/32Descriptor(isPtr)` | |
| `uint64` | `bigint` | `NewUint64Descriptor(isPtr)` | |
| `bool` | `boolean` | `NewBoolDescriptor(isPtr)` | |
| `string` | `string` | `NewStringDescriptor(isPtr)` | UTF-8 |
| `bytes` | `Uint8Array` | `NewBytesDescriptor(isPtr)` | |
| `float` | `number` | `NewFloatDescriptor(isPtr)` | float32 精度（`Math.fround`） |
| `double` | `number` | `NewDoubleDescriptor(isPtr)` | |
| `date` | `Date \| null` | `NewTimeDescriptor(isPtr)` | **无论可空与否类型都含 null**（零时间用 null 表示） |
| `decimal` | `string` | `NewDecimalDescriptor(isPtr)` | 十进制字符串，`"0"` 为零值 |
| data 引用 `D` | `DTs`（可空则 `DTs \| null`） | `NewRefDescriptor(dTsDesc, isPtr)` | |
| `T[]` | `T[]` / `(T \| null)[]` | `NewListDescriptor(innerDesc, isPtr)` | 内层 isPtr = 元素是否可空 |
| `V<K>` | `Map<bigint, VTs>` / `Map<string, VTs>` | `NewMapDescriptor(kDesc, vDesc, isPtr)` | K=`int64`→bigint 键；`string`→string 键 |

**isPtr 判定规则**（与 Go 指针语义一一对应）：

| IDL 写法 | 外层 isPtr | 内层（元素/值）isPtr |
|---|---|---|
| `T` | `false` | — |
| `T?` | `true` | — |
| `T[]` | `false` | `false` |
| `T?[]` | `false` | `true` |
| `V<K>` | `false` | `false` |
| `V?<K>` | `false` | `true` |
| 引用 `D` / `D?` | 见字段 | `NewRefDescriptor(父desc, isPtr)` |

## 5. 生成规则

### 5.1 命名

- 类名 = IDL data 名 + `Ts` 后缀（如 `HBufSubTest` → `HBufSubTestTs`）。
- 属性名 = IDL 字段名转 **camelCase 首字母小写**（`V1` → `v1`；已是小写驼峰则保持不变）。
- descriptor 单例名 = data 名 camelCase 首字母小写 + `Desc`（`HBufSubTest` → `hbufSubTestDesc`）。

### 5.2 默认值初始化

| TS 类型 | 默认值 |
|---|---|
| `number` | `= 0` |
| `bigint` | `= 0n` |
| `boolean` | `= false` |
| `string` | `= ""` |
| `Uint8Array` | `= new Uint8Array(0)` |
| `Date \| null` | `= null` |
| `T[]` | `= []` |
| `Map<K,V>` | `= new Map()` |
| 可空类型 | `= null` |
| data 引用（非可空） | `= new XTs()`（零值实例） |

### 5.3 声明顺序（重要）

1. **被引用的 data 的 desc 必须先声明**：`NewDataDescriptor` 构造时立即求值 `fields` Map 中的 desc 引用，依赖的 desc 常量必须在前。
2. **递归 data 必须用 `NewSyncDescriptor` 包裹**（延迟构建，运行时才求值闭包）：
   ```ts
   const nodeDesc: Descriptor = NewSyncDescriptor(() => NewDataDescriptor(
       () => new NodeTs(),
       new Map<number, FieldDesc>([
           [1, {desc: NewListDescriptor(NewRefDescriptor(nodeDesc, true), false), ...}],
       ]),
   ))
   ```
3. 类在 desc 之后声明；类内 `descriptors()` 返回该 desc 单例。

### 5.4 输出文件组织

- 每个 `.hbuf` 文件生成一个 `.ts` 文件（或按 data 拆分）。
- 生成文件只 import 运行时（建议 `import {HData, Descriptor, FieldDesc, ...} from "<运行时包>"`）。
- 可选：附 `buildXxx()` 填值函数（供测试）。

## 6. 生成代码模板

### 6.1 简单 data（无引用、无递归）

IDL：
```
data User {
    int64 Id = 1
    string? Name = 2
    bool Active = 3
}
```

生成代码：
```ts
import {DataDescriptor, Descriptor, FieldDesc, HData, NewBoolDescriptor, NewDataDescriptor,
    NewInt64Descriptor, NewStringDescriptor} from "../src"

class UserTs implements HData {
    id: bigint = 0n
    name: string | null = null
    active: boolean = false

    descriptors(): Descriptor {
        return userDesc
    }
}

const userDesc: DataDescriptor = NewDataDescriptor(
    () => new UserTs(),
    new Map<number, FieldDesc>([
        [1, {desc: NewInt64Descriptor(), get: (o: UserTs) => o.id, set: (o: UserTs, v) => o.id = v}],
        [2, {desc: NewStringDescriptor(true), get: (o: UserTs) => o.name, set: (o: UserTs, v) => o.name = v}],
        [3, {desc: NewBoolDescriptor(), get: (o: UserTs) => o.active, set: (o: UserTs, v) => o.active = v}],
    ]),
)
```

### 6.2 引用其他 data

```ts
class OrderTs implements HData {
    user: UserTs = new UserTs()          // 非可空引用：零值实例
    tags: string[] = []
    userOpt: UserTs | null = null        // 可空引用
    attrs: Map<string, string> = new Map()

    descriptors(): Descriptor {
        return orderDesc
    }
}

const orderDesc: DataDescriptor = NewDataDescriptor(
    () => new OrderTs(),
    new Map<number, FieldDesc>([
        [1, {desc: NewRefDescriptor(userDesc), get: (o: OrderTs) => o.user, set: (o: OrderTs, v) => o.user = v}],
        [2, {desc: NewListDescriptor(NewStringDescriptor(), false), get: (o: OrderTs) => o.tags, set: (o: OrderTs, v) => o.tags = v}],
        [3, {desc: NewRefDescriptor(userDesc, true), get: (o: OrderTs) => o.userOpt, set: (o: OrderTs, v) => o.userOpt = v}],
        [4, {desc: NewMapDescriptor(NewStringDescriptor(), NewStringDescriptor(), false), get: (o: OrderTs) => o.attrs, set: (o: OrderTs, v) => o.attrs = v}],
    ]),
)
```

### 6.3 递归 data

```ts
const nodeDesc: Descriptor = NewSyncDescriptor(() => NewDataDescriptor(
    () => new NodeTs(),
    new Map<number, FieldDesc>([
        [1, {desc: NewListDescriptor(NewRefDescriptor(nodeDesc, true), false),
             get: (o: NodeTs) => o.children, set: (o: NodeTs, v) => o.children = v}],
        [2, {desc: NewRefDescriptor(nodeDesc, true),
             get: (o: NodeTs) => o.parent, set: (o: NodeTs, v) => o.parent = v}],
    ]),
))

class NodeTs implements HData {
    children: (NodeTs | null)[] = []
    parent: NodeTs | null = null

    descriptors(): Descriptor {
        return nodeDesc
    }
}
```

### 6.4 extends（父类型内嵌）

IDL：`data HBufSub1Test:HBufBTest = 1, HBufSubTest = 2, HBufTest = 3 { ... }`

对应 Go 生成方式：父类型为**内嵌字段**，挂入 `extends_` Map（键 = 扩展 ID）：

```ts
class HBufSub1TestTs implements HData {
    hbufBTest: HBufBTestTs = new HBufBTestTs()      // 内嵌父类型，属性名 = 父类型名 camelCase
    hbufSubTest: HBufSubTestTs = new HBufSubTestTs()
    hbufTest: HBufTestTs = new HBufTestTs()
    v1: bigint = 0n
    v2: bigint | null = null

    descriptors(): Descriptor {
        return hbufSub1TestDesc
    }
}

const hbufSub1TestDesc: DataDescriptor = NewDataDescriptor(
    () => new HBufSub1TestTs(),
    new Map<number, FieldDesc>([
        [1, {desc: NewInt64Descriptor(), get: (o: HBufSub1TestTs) => o.v1, set: (o: HBufSub1TestTs, v) => o.v1 = v}],
        [2, {desc: NewInt64Descriptor(true), get: (o: HBufSub1TestTs) => o.v2, set: (o: HBufSub1TestTs, v) => o.v2 = v}],
    ]),
    new Map<number, FieldDesc>([   // extends_：父类型按扩展 ID
        [1, {desc: NewRefDescriptor(hbufBTestDesc), get: (o: HBufSub1TestTs) => o.hbufBTest, set: (o: HBufSub1TestTs, v) => o.hbufBTest = v}],
        [2, {desc: NewRefDescriptor(hbufSubTestDesc), get: (o: HBufSub1TestTs) => o.hbufSubTest, set: (o: HBufSub1TestTs, v) => o.hbufSubTest = v}],
        [3, {desc: NewRefDescriptor(hbufTestDesc), get: (o: HBufSub1TestTs) => o.hbufTest, set: (o: HBufSub1TestTs, v) => o.hbufTest = v}],
    ]),
)
```

## 7. 协议语义速查（生成器无需实现，但必须理解）

- **字段省略**：字段值 `isEmpty` 时不写入（零值 / null）；可空字段**非 null 恒写入**（即使值为零值）。唯一例外：`uint8` 值 `0` 恒省略（与 Go 一致）。
- **元素编码**（List 元素 / Map Key/Value）：无 ID 字节；空值写 `TAG(类型, idLen=0, 1)` 单字节空标记保证数量对齐；`uint8` 零值元素为 `TAG(TUint, idLen=2, 1)` 特例。
- **非可空 data 字段零值**仍写入 3 字节空结构（Go `IsEmpty` 只判 nil）。
- **date**：固定 9 字节（42 位秒 + 30 位纳秒），JS `Date` 毫秒精度（纳秒为毫秒的整数倍时才与 Go 字节级一致）。
- **decimal**：字符串形式编码（`Number(v) === 0` 判空）。
- **字段按 ID 升序**编码（`DataDescriptor` 内部已排序）；ID > 255 自动用 2 字节。
- **跨类型解码**：数值类字段可接受 `TBool/TInt/TUint/TFloat/TBytes` 输入；**未知字段自动跳过**（含嵌套 data 的 EXTENDS-BLOCK 递归跳过）。
- **float/double**：位模式编码（`Float32Array/Uint32Array` 共享 buffer），`NaN/±Inf` 保持；字段级 `-0` 因 `v === 0` 判空省略，元素级 `-0` 按位模式保真。

## 8. RPC 集成（可选）

生成的数据类可直接接入 hrpc 可插拔编解码器：

```ts
import {Marshal, Unmarshal, HData} from "<运行时包>"

// 编码器：Data → ArrayBuffer
export function NewHBufEncode(): Encoder {
    return (v: Data): ArrayBuffer => Marshal(v as unknown as HData).slice().buffer
}

// 解码器：ArrayBuffer → Data（from 工厂创建实例，Unmarshal 就地填充）
export function NewHBufDecoder(): Decoder {
    return (buffer: ArrayBuffer, from: FromMap, tag: string): Data => {
        const data = from({} as Record<string, any>, tag) as unknown as HData
        Unmarshal(new Uint8Array(buffer), data)
        return data as unknown as Data
    }
}
```

## 9. 测试与验证规范（golden 方法）

生成器的产出必须按以下方式验证字节级兼容：

1. **Go 侧**写临时测试，用 Go 版 `Marshal` 对样例数据编码，输出 hex 到 golden 文件。
2. **TS 侧**生成相同语义的填值 + 断言 `Marshal(tsData) 的 hex === golden`（字节级）、`Unmarshal(golden, dst)` 后逐字段语义一致（语义级）、round-trip。
3. **Go map 字节序随机**：多元素 map 只做长度对比 + 解码语义对比；字节级对比用**单元素 map**。
4. **时间精度陷阱**：字节级对比的时间值必须用**整毫秒**（纳秒为 1e6 整数倍）。
5. **NaN**：使用规范位模式（float32 `0x7FC00000`、float64 `0x7FF8000000000000`）才能字节级一致。
6. 边界用例清单（至少覆盖）：2 字节 ID（>255）、变长整数 1~8 字节转折点、大 count（>255 元素）、深层递归嵌套、nil 元素、空串 key、uint8 零值特例、NaN/±Inf/-0。

参考实现：`test/hbuf.test.ts`（基础协议 12 例）、`test/hbuf128.test.ts`（128 字段完整数据）、`test/hbuf_edge.test.ts`（边界 9 例）。

## 10. 陷阱清单（生成器实现时必须遵守）

1. **desc 声明顺序**：被引用 desc 必须先声明；递归必须 `NewSyncDescriptor`（否则运行时 TDZ 报错）。
2. **`date` 类型恒为 `Date | null`**（即使 IDL 不可空）——零时间用 null 表示，不能声明为 `Date`。
3. **`int64/uint64` 必须用 `bigint`**（`number` 会在 2^53 以上丢精度，与 Go 字节级不一致）。
4. **非可空 data 引用默认值**是 `new XTs()` 而非 null（零值实例会编码为 3 字节空结构）。
5. **可空字段值判定**：`v == null` 是空；`false`/`""`/`new Uint8Array(0)` 等非 null 零值**仍写入**。
6. **uint8 值 0 恒省略**（包括可空 uint8 指向 0 时），这是与 Go 一致的保留怪癖。
7. **`Uint8Array` 是引用类型**：字段默认值必须用 `= new Uint8Array(0)` 而非类属性声明共享。
8. **不要给生成代码添加任何序列化无关逻辑**（如 toJSON、构造函数参数），保持与 Go 生成代码结构对称。
9. **Map 键序**：JS `Map` 按插入序编码（Go map 随机序），字节级对比只能用单元素 map；解码端不受序影响。
10. **空标记容器元素解码为 null**：`[][]byte` / `[]T[]` / `[]Map` 等嵌套容器中的空标记元素与 Go 零值一致返回 null（不是 `[]`/`new Map()`）；生成类型须用 `(Uint8Array | null)[]` 等可空声明。
11. **count==0 的容器解码不触碰字段**：list/map 的 count==0 与 Go 一致保持字段原值（运行时返回 undefined 跳过 set）；非指针 data 同理；指针 data（isPtr）每次解码都**分配新实例**覆盖旧引用（Go 行为）。
12. **可空容器字段（`*[]T`/`*Map<K,V>`）**：解码 count>0 时分配新实例；count==0 保持 nil（Go 端 List/MapDescriptor 已修复同样行为，两端一致）。
13. **值 data 元素（非指针列表元素）零值编码为 3 字节空结构**（`TAG(TData,1,1) | 00 | EXTENDS`），与 Go `p != nil` 值元素一致；不要误写成空标记。
