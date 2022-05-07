import 'dart:typed_data';
import 'dart:convert';
import 'package:hbuf_dart/hbuf_dart.dart';

abstract class Base implements Data{
  /// 姓名
  List<String?>? get namePeople;

  set namePeople(List<String?>? value);

  factory Base({
    List<String?>? namePeople,
  }){
    return _Base(
      namePeople: namePeople,
    );
  }

  static Base fromMap(Map<String, dynamic> map){
    return _Base.fromMap(map);
  }

  static Base fromData(ByteData data){
    return _Base.fromData(data);
  }

}

class _Base implements Base {
  @override
  List<String?>? namePeople;

  _Base({
    this.namePeople,
  });

  static _Base fromMap(Map<String, dynamic> map){
    return _Base(
      namePeople: map["name_people"],
    );
  }

  @override
  Map<String, dynamic> toMap() {
    return {
      "name_people": namePeople,
    };
  }
  static _Base fromData(ByteData data){
    return _Base(
    );
  }

  @override
  ByteData toData() {
    return ByteData.view(Uint8List(12).buffer);
  }
}

abstract class People implements Data, Base{
  /// 姓名
  @override
  List<String?>? get namePeople;

  @override
  set namePeople(List<String?>? value);

  /// 姓名
  Map<String, String?>? get map;

  set map(Map<String, String?>? value);

  factory People({
    List<String?>? namePeople,
    Map<String, String?>? map,
  }){
    return _People(
      namePeople: namePeople,
      map: map,
    );
  }

  static People fromMap(Map<String, dynamic> map){
    return _People.fromMap(map);
  }

  static People fromData(ByteData data){
    return _People.fromData(data);
  }

}

class _People implements People {
  @override
  List<String?>? namePeople;

  @override
  Map<String, String?>? map;

  _People({
    this.namePeople,
    this.map,
  });

  static _People fromMap(Map<String, dynamic> map){
    return _People(
      namePeople: map["name_people"],
      map: map["map"],
    );
  }

  @override
  Map<String, dynamic> toMap() {
    return {
      "name_people": namePeople,
      "map": map,
    };
  }
  static _People fromData(ByteData data){
    return _People(
    );
  }

  @override
  ByteData toData() {
    return ByteData.view(Uint8List(12).buffer);
  }
}

abstract class PeopleServer{
  /// 获得年龄
  Future<People> getPeople(People userId, [Context? ctx]);

  /// 获得姓名
  Future<People> getAddress(People userId, [Context? ctx]);

}

class PeopleServerImp extends ServerClient implements PeopleServer{
  PeopleServerImp(Client client):super(client);

  @override
  String get name => "PeopleServer";

  @override
  int get id => 1;

  @override
  Future<People> getPeople(People userId, [Context? ctx]){
    return invoke<People>("PeopleServer/getPeople", 1 << 32 | 1, userId, People.fromMap, People.fromData);
  }

  @override
  Future<People> getAddress(People userId, [Context? ctx]){
    return invoke<People>("PeopleServer/getAddress", 1 << 32 | 2, userId, People.fromMap, People.fromData);
  }

}

class PeopleServerRouter extends ServerRouter{
  final PeopleServer server;

  @override
  String get name => "PeopleServer";

  @override
  int get id => 1;

  Map<String, ServerInvoke> _invokeNames = {};

  Map<int, ServerInvoke> _invokeIds = {};

  @override
  Map<String, ServerInvoke> get invokeNames => _invokeNames;

  @override
  Map<int, ServerInvoke> get invokeIds => _invokeIds;

  PeopleServerRouter(this.server){
    _invokeNames = {
      "PeopleServer/getPeople": ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromMap(json.decode(utf8.decode(buf)));
        },
        formData: (Data data) async {
     	   return utf8.encode(json.encode(data.toMap()));
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getPeople(data as People, ctx);
        },
      ),
      "PeopleServer/getAddress": ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromMap(json.decode(utf8.decode(buf)));
        },
        formData: (Data data) async {
     	   return utf8.encode(json.encode(data.toMap()));
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getAddress(data as People, ctx);
        },
      ),
    };

    _invokeIds = {
        1 << 32 | 1: ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromData(ByteData.view(Uint8List.fromList(buf).buffer));
        },
        formData: (Data data) async {
     	   return data.toData().buffer.asUint8List();
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getPeople(data as People, ctx);
        },
      ),
        1 << 32 | 2: ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromData(ByteData.view(Uint8List.fromList(buf).buffer));
        },
        formData: (Data data) async {
     	   return data.toData().buffer.asUint8List();
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getAddress(data as People, ctx);
        },
      ),
    };

  }

}

abstract class StudentServer implements PeopleServer{
  /// 获得年龄
  @override
  Future<People> getPeople(People userId, [Context? ctx]);

  /// 获得姓名
  Future<People> getNumber(People userId, [Context? ctx]);

}

class StudentServerImp extends ServerClient implements StudentServer{
  StudentServerImp(Client client):super(client);

  @override
  String get name => "StudentServer";

  @override
  int get id => 2;

  @override
  Future<People> getPeople(People userId, [Context? ctx]){
    return invoke<People>("StudentServer/getPeople", 2 << 32 | 1, userId, People.fromMap, People.fromData);
  }

  @override
  Future<People> getNumber(People userId, [Context? ctx]){
    return invoke<People>("StudentServer/getNumber", 2 << 32 | 2, userId, People.fromMap, People.fromData);
  }

  @override
  Future<People> getAddress(People userId, [Context? ctx]){
    return invoke<People>("PeopleServer/getAddress", 1 << 32 | 2, userId, People.fromMap, People.fromData);
  }

}

class StudentServerRouter extends ServerRouter{
  final StudentServer server;

  @override
  String get name => "StudentServer";

  @override
  int get id => 2;

  Map<String, ServerInvoke> _invokeNames = {};

  Map<int, ServerInvoke> _invokeIds = {};

  @override
  Map<String, ServerInvoke> get invokeNames => _invokeNames;

  @override
  Map<int, ServerInvoke> get invokeIds => _invokeIds;

  StudentServerRouter(this.server){
    _invokeNames = {
      "StudentServer/getPeople": ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromMap(json.decode(utf8.decode(buf)));
        },
        formData: (Data data) async {
     	   return utf8.encode(json.encode(data.toMap()));
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getPeople(data as People, ctx);
        },
      ),
      "StudentServer/getNumber": ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromMap(json.decode(utf8.decode(buf)));
        },
        formData: (Data data) async {
     	   return utf8.encode(json.encode(data.toMap()));
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getNumber(data as People, ctx);
        },
      ),
      "PeopleServer/getAddress": ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromMap(json.decode(utf8.decode(buf)));
        },
        formData: (Data data) async {
     	   return utf8.encode(json.encode(data.toMap()));
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getAddress(data as People, ctx);
        },
      ),
    };

    _invokeIds = {
        2 << 32 | 1: ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromData(ByteData.view(Uint8List.fromList(buf).buffer));
        },
        formData: (Data data) async {
     	   return data.toData().buffer.asUint8List();
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getPeople(data as People, ctx);
        },
      ),
        2 << 32 | 2: ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromData(ByteData.view(Uint8List.fromList(buf).buffer));
        },
        formData: (Data data) async {
     	   return data.toData().buffer.asUint8List();
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getNumber(data as People, ctx);
        },
      ),
        1 << 32 | 2: ServerInvoke(
        toData: (List<int> buf) async {
          return People.fromData(ByteData.view(Uint8List.fromList(buf).buffer));
        },
        formData: (Data data) async {
     	   return data.toData().buffer.asUint8List();
        },
        invoke: (Context ctx, Data data) async {
     	   return await server.getAddress(data as People, ctx);
        },
      ),
    };

  }

}

