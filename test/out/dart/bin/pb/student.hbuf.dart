import 'package:hbuf_dart/hbuf_dart.dart';
import "people.hbuf.dart";

class Gender{
  final int value;
  final String name;

  const Gender._(this.value, this.name);

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is Gender &&
          runtimeType == other.runtimeType &&
          value == other.value;

  @override
  int get hashCode => value.hashCode;

  static Gender valueOf(int value) {
  	for (var item in values) {
  		if (item.value == value) {
  			return item;
  		}
  	}
  	throw 'Get Gender by value error, value=$value';
  }

  static Gender nameOf(String name) {
  	for (var item in values) {
  		if (item.name == name) {
  			return item;
  		}
  	}
  	throw 'Get Gender by name error, name=$name';
  }

  static const Girl = Gender._(1, 'Girl');
  static const Boy = Gender._(2, 'Boy');

  static const List<Gender> values = [
    Girl,
    Boy,
  ];

}

abstract class StudentData implements People, Data {
  /// 编号
  int? no;

  Gender? gender;

  List<int?>? ll;

  Map<int, String?>? bookTags;

  factory StudentData({
    int? no,
    Gender? gender,
    List<int?>? ll,
    Map<int, String?>? bookTags,
    String? name,
    int? age,
  }){
    return _StudentData(
      no: no,
      gender: gender,
      ll: ll,
      bookTags: bookTags,
      name: name,
      age: age,
    );
  }

  static StudentData? fromMap(Map<String, dynamic> map){
    return _StudentData.fromMap(map);
  }

}

class _StudentData implements StudentData {
  @override
  int? no;

  @override
  Gender? gender;

  @override
  List<int?>? ll;

  @override
  Map<int, String?>? bookTags;

  @override
  String? name;

  @override
  int? age;

  _StudentData({
    this.no,
    this.gender,
    this.ll,
    this.bookTags,
    this.name,
    this.age,
  });

  static _StudentData? fromMap(Map<String, dynamic> map){
    return _StudentData(
      no: map["no"],
      gender: map["gender"],
      ll: map["ll"],
      bookTags: map["book_tags"],
      name: map["name"],
      age: map["age"],
    );
  }

  @override
  Map<String, dynamic> toMap() {
    return {
      "no": no,
      "gender": gender,
      "ll": ll,
      "book_tags": bookTags,
      "name": name,
      "age": age,
    };
  }
}
