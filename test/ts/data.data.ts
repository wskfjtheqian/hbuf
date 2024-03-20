import * as $4 from "./data.data"
import * as $3 from "./data.enum"
import * as d from "decimal.js"
import h from "hbuf_ts"
import Long from "long"

///12
export class GetInfoReq implements h.Data {
	userId: Long | null = null;

	name: string | null = null;

	age: number | null = null;

	public static fromJson(json: Record<string, any>): GetInfoReq{
		let ret = new GetInfoReq()
		let temp:any
		ret.userId = null == (temp = json["user_id"]) ? null : Long.fromString(temp)
		ret.name = null == (temp = json["name"]) ? null : temp.toString()
		ret.age = null == (temp = json["age"]) ? null : Number(temp).valueOf()
		return ret
	}


	public toJson(): Record<string, any> {
		return {
			"user_id": this.userId?.toString(),
			"name": this.name,
			"age": this.age,
		};
	}

	public static fromData(data: BinaryData): GetInfoReq {
		let ret = new GetInfoReq()
		return ret
	}

	public toData(): BinaryData {
		return new ArrayBuffer(0)
	}

}

export class InfoReq implements h.Data {
	userId: Long | null = null;

	public static fromJson(json: Record<string, any>): InfoReq{
		let ret = new InfoReq()
		let temp:any
		ret.userId = null == (temp = json["user_id"]) ? null : Long.fromString(temp)
		return ret
	}


	public toJson(): Record<string, any> {
		return {
			"user_id": this.userId?.toString(),
		};
	}

	public static fromData(data: BinaryData): InfoReq {
		let ret = new InfoReq()
		return ret
	}

	public toData(): BinaryData {
		return new ArrayBuffer(0)
	}

}

export class InfoSet implements h.Data {
	userId: Long | null = null;

	name: string | null = null;

	age: number | null = null;

	public static fromJson(json: Record<string, any>): InfoSet{
		let ret = new InfoSet()
		let temp:any
		ret.userId = null == (temp = json["user_id"]) ? null : Long.fromString(temp)
		ret.name = null == (temp = json["name"]) ? null : temp.toString()
		ret.age = null == (temp = json["age"]) ? null : Number(temp).valueOf()
		return ret
	}


	public toJson(): Record<string, any> {
		return {
			"user_id": this.userId?.toString(),
			"name": this.name,
			"age": this.age,
		};
	}

	public static fromData(data: BinaryData): InfoSet {
		let ret = new InfoSet()
		return ret
	}

	public toData(): BinaryData {
		return new ArrayBuffer(0)
	}

}

///12
export class GetInfoResp implements h.Data {
	v1: number = 0;

	b1: number | null = null;

	v2: number = 0;

	b2: number | null = null;

	v3: number = 0;

	b3: number | null = null;

	v4: Long = Long.ZERO;

	b4: Long | null = null;

	v5: number = 0;

	b5: number | null = null;

	v6: number = 0;

	b6: number | null = null;

	v7: number = 0;

	b7: number | null = null;

	v8: Long = Long.ZERO;

	b8: Long | null = null;

	v9: boolean = false;

	b9: boolean | null = null;

	v10: number = 0.0;

	b10: number | null = null;

	v11: number = 0.0;

	b11: number | null = null;

	v12: string = "";

	b12: string | null = null;

	v13: Date = new Date();

	b13: Date | null = null;

	v14: d.Decimal = new d.Decimal(0);

	b14: d.Decimal | null = null;

	v15: $3.Status = $3.Status.valueOf(0);

	b15: $3.Status | null = null;

	v16: $4.GetInfoReq = new $4.GetInfoReq();

	b16: $4.GetInfoReq | null = null;

	v17: ($4.GetInfoReq)[] = [];

	b17: ($4.GetInfoReq)[] | null = null;

	v19: ($4.GetInfoReq | null)[] = [];

	b19: ($4.GetInfoReq | null)[] | null = null;

	v18: Record<(number), ($4.GetInfoReq)> = {};

	b18: Record<(string), ($4.GetInfoReq)> | null = null;

	v20: Record<(string), ($4.GetInfoReq | null)> = {};

	b20: Record<(string), ($4.GetInfoReq | null)> | null = null;

	public static fromJson(json: Record<string, any>): GetInfoResp{
		let ret = new GetInfoResp()
		let temp:any
		ret.v1 = null == (temp = json["v1"]) ? 0 : (Number(temp).valueOf() || 0)
		ret.b1 = null == (temp = json["b1"]) ? null : Number(temp).valueOf()
		ret.v2 = null == (temp = json["v2"]) ? 0 : (Number(temp).valueOf() || 0)
		ret.b2 = null == (temp = json["b2"]) ? null : Number(temp).valueOf()
		ret.v3 = null == (temp = json["v3"]) ? 0 : (Number(temp).valueOf() || 0)
		ret.b3 = null == (temp = json["b3"]) ? null : Number(temp).valueOf()
		ret.v4 = null == (temp = json["v4"]) ? Long.ZERO : Long.fromString(temp)
		ret.b4 = null == (temp = json["b4"]) ? null : Long.fromString(temp)
		ret.v5 = null == (temp = json["v5"]) ? 0 : (Number(temp).valueOf() || 0)
		ret.b5 = null == (temp = json["b5"]) ? null : Number(temp).valueOf()
		ret.v6 = null == (temp = json["v6"]) ? 0 : (Number(temp).valueOf() || 0)
		ret.b6 = null == (temp = json["b6"]) ? null : Number(temp).valueOf()
		ret.v7 = null == (temp = json["v7"]) ? 0 : (Number(temp).valueOf() || 0)
		ret.b7 = null == (temp = json["b7"]) ? null : Number(temp).valueOf()
		ret.v8 = null == (temp = json["v8"]) ? Long.ZERO : Long.fromString(temp)
		ret.b8 = null == (temp = json["b8"]) ? null : Long.fromString(temp)
		ret.v9 = null == (temp = json["v9"]) ? false : ("true" === temp ? true : Boolean(temp))
		ret.b9 = null == (temp = json["b9"]) ? null : ("true" === temp ? true : Boolean(temp))
		ret.v10 = null == (temp = json["v10"]) ? 0 : (Number(temp).valueOf() || 0)
		ret.b10 = null == (temp = json["b10"]) ? null : Number(temp).valueOf()
		ret.v11 = null == (temp = json["v11"]) ? 0 : (Number(temp).valueOf() || 0)
		ret.b11 = null == (temp = json["b11"]) ? null : Number(temp).valueOf()
		ret.v12 = null == (temp = json["v12"]) ? "" : temp.toString()
		ret.b12 = null == (temp = json["b12"]) ? null : temp.toString()
		ret.v13 = null == (temp = json["v13"]) ? new Date(0): new Date(temp)
		ret.b13 = null == (temp = json["b13"]) ? null : new Date(temp)
		ret.v14 = null == (temp = json["v14"]) ? new d.Decimal(0) : new d.Decimal(temp) 
		ret.b14 = null == (temp = json["b14"]) ? null : new d.Decimal(temp)
		ret.v15 = null == (temp = json["v15"]) ? $3.Status.valueOf(0) : $3.Status.valueOf(Number(temp).valueOf())
		ret.b15 = null == (temp = json["b15"]) ? null : $3.Status.valueOf(Number(temp).valueOf())
		ret.v16 = null == (temp = json["v16"]) ? $4.GetInfoReq.fromJson({}) : $4.GetInfoReq.fromJson(temp)
		ret.b16 = null == (temp = json["b16"]) ? null : $4.GetInfoReq.fromJson(temp)
		ret.v17 = null == (temp = json["v17"]) ? [] : (h.isArray(temp) ? [] : (h.convertArray(temp, (item) => null == item ? $4.GetInfoReq.fromJson({}) : $4.GetInfoReq.fromJson(item))))
		ret.b17 = null == (temp = json["b17"]) ? null : (h.isArray(temp) ? null : (h.convertArray(temp, (item) => null == item ? $4.GetInfoReq.fromJson({}) : $4.GetInfoReq.fromJson(item))))
		ret.v19 = null == (temp = json["v19"]) ? [] : (h.isArray(temp) ? [] : (h.convertArray(temp, (item) => null == item ? null : $4.GetInfoReq.fromJson(item))))
		ret.b19 = null == (temp = json["b19"]) ? null : (h.isArray(temp) ? null : (h.convertArray(temp, (item) => null == item ? null : $4.GetInfoReq.fromJson(item))))
		ret.v18 = null == (temp = json["v18"]) ? {} : (h.isRecord(temp) ? {} : (h.convertRecord(temp, (key, value) => new h.RecordEntry(null == key ? $3.Status.valueOf(0) : $3.Status.valueOf(Number(key).valueOf()),null == value ? $4.GetInfoReq.fromJson({}) : $4.GetInfoReq.fromJson(value)))))
		ret.b18 = null == (temp = json["b18"]) ? null : (h.isRecord(temp) ? null : (h.convertRecord(temp, (key, value) => new h.RecordEntry(null == key ? "" : key.toString(),null == value ? $4.GetInfoReq.fromJson({}) : $4.GetInfoReq.fromJson(value)))))
		ret.v20 = null == (temp = json["v20"]) ? {} : (h.isRecord(temp) ? {} : (h.convertRecord(temp, (key, value) => new h.RecordEntry(null == key ? "" : key.toString(),null == value ? null : $4.GetInfoReq.fromJson(value)))))
		ret.b20 = null == (temp = json["b20"]) ? null : (h.isRecord(temp) ? null : (h.convertRecord(temp, (key, value) => new h.RecordEntry(null == key ? "" : key.toString(),null == value ? null : $4.GetInfoReq.fromJson(value)))))
		return ret
	}


	public toJson(): Record<string, any> {
		return {
			"v1": this.v1,
			"b1": this.b1,
			"v2": this.v2,
			"b2": this.b2,
			"v3": this.v3,
			"b3": this.b3,
			"v4": this.v4.toString(),
			"b4": this.b4?.toString(),
			"v5": this.v5,
			"b5": this.b5,
			"v6": this.v6,
			"b6": this.b6,
			"v7": this.v7,
			"b7": this.b7,
			"v8": this.v8.toString(),
			"b8": this.b8?.toString(),
			"v9": this.v9,
			"b9": this.b9,
			"v10": this.v10,
			"b10": this.b10,
			"v11": this.v11,
			"b11": this.b11,
			"v12": this.v12,
			"b12": this.b12,
			"v13": this.v13.getTime(),
			"b13": this.b13?.getTime(),
			"v14": this.v14.toString(),
			"b14": this.b14?.toString(),
			"v15": this.v15.value,
			"b15": this.b15?.value,
			"v16": this.v16.toJson(),
			"b16": this.b16?.toJson(),
			"v17": this.v17.map((e) => e.toJson()),
			"b17": this.b17?.map((e) => e.toJson()),
			"v19": this.v19.map((e) => e?.toJson()),
			"b19": this.b19?.map((e) => e?.toJson()),
			"v18": this.v18.map((key,value) => MapEntry(key.value,value.value)),
			"b18": this.b18?.map((key,value) => MapEntry(key,value),
			"v20": this.v20.map((key,value) => MapEntry(key,value)),
			"b20": this.b20?.map((key,value) => MapEntry(key,value),
		};
	}

	public static fromData(data: BinaryData): GetInfoResp {
		let ret = new GetInfoResp()
		return ret
	}

	public toData(): BinaryData {
		return new ArrayBuffer(0)
	}

}

