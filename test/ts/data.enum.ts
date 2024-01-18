
///12
export class Status{
	public readonly value: number

	public readonly name: string

	private constructor(value: number, name: string) {
		this.value = value;
		this.name = name;
	}
	public static valueOf(value: number): Status {
		for (var i in Status.values) {
			if (Status.values[i].value == value) {
				return Status.values[i];
			}
		}
		throw 'Get Status by value error, value=${value}';
	}

	public static nameOf(name: string): Status {
		for (var i in Status.values) {
			if (Status.values[i].name == name) {
				return Status.values[i];
			}
		}
		throw 'Get Status by name error, name=${name}';
	}

	///启用
	public static readonly ENABLE = new Status(0, "Enable");

	///禁用
	public static readonly DISABLED = new Status(1, "Disabled");


	public static readonly values:Status[] = [
		Status.ENABLE,
		Status.DISABLED,
	];

	toString():string {
		return this.name;
	}

}
