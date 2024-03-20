import * as $1 from "./data.data"
import * as $2 from "./data.server"
import h from "hbuf_ts"

///12
export interface UserServer {
	//12
	getInfo(req: $1.GetInfoReq, ctx?: h.Context): Promise<$1.GetInfoResp>

}

export class UserServerClient extends h.ServerClient implements $2.UserServer{
	constructor(client: h.Client){
		super(client)
	}
	get name(): string {
		return "user_server"
	}

	get id(): number {
		return 0	
	}

	//12
	getInfo(req: $1.GetInfoReq, ctx?: h.Context): Promise<$1.GetInfoResp> {
		return this.invoke<$1.GetInfoResp>("user_server/get_info", 0 << 32 | 0, req, $1.GetInfoResp.fromJson, $1.GetInfoResp.fromData);
	}

}

export class UserServerRouter implements h.ServerRouter {
	readonly server: UserServer

	invoke: Record<string, h.ServerInvoke>

	getInvoke(): Record<string, h.ServerInvoke> {
		return this.invoke
	}

	getName(): string {
		return "user_server"
	}

	getId(): number {
		return 0
	}

	constructor(server: UserServer) {
		this.server = server
		this.invoke = {
			"user_server/get_info": {
				formData(data: BinaryData | Record<string, any>): h.Data {
					return $1.GetInfoReq.fromJson(data)
				},
				toData(data: h.Data): BinaryData | Record<string, any> {
					return data.toJson()
				},
				invoke(data: h.Data, ctx?: h.Context): Promise<h.Data | void> {
					return server.getInfo(data as $1.GetInfoReq, ctx);
				}
			},
		}
	}
}
