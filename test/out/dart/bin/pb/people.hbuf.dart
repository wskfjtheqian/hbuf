import 'package:hbuf_dart/hbuf_dart.dart';

abstract class People implements Data {
  /// 姓名
  String? name;

  /// 年龄
  int? age;

  factory People({
    String? name,
    int? age,
  }){
    return _People(
      name: name,
      age: age,
    );
  }

  static People? fromMap(Map<String, dynamic> map){
    return _People.fromMap(map);
  }

}

class _People implements People {
  @override
  String? name;

  @override
  int? age;

  _People({
    this.name,
    this.age,
  });

  static _People? fromMap(Map<String, dynamic> map){
    return _People(
      name: map["name"],
      age: map["age"],
    );
  }

  @override
  Map<String, dynamic> toMap() {
    return {
      "name": name,
      "age": age,
    };
  }
}
