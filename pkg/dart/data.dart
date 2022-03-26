enum AA {
  B,
  C,
}

class Gender {
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
    throw "Get Gender by value error, value=$value";
  }

  static Gender nameOf(String name) {
    for (var item in values) {
      if (item.name == name) {
        return item;
      }
    }
    throw "Get Gender by name error, name=$name";
  }

  static const girl = Gender._(0, "girl");
  static const boy = Gender._(0, "boy");

  static const List<Gender> values = [
    girl,
    boy,
  ];
}

//
// abstract class Student implements Data {
//
//   /// 姓名
//   String? Name;
//
//   List<String>? Info;
//
//   /// 姓名
//   Map<String, int>? other;
//
//   static Student create({String? Name, List<String>? Info, Map<String, int>? other, }){
//     return _Student(Name: Name, Info: Info, other: other, );
//   }
//
//   static Student fromMap(Map<String, dynamic> map){
//     return _Student.fromMap(map);
//   }
//
// }
//
// class _Student implements Student {
//
//   @override
//   String? Name;
//
//   @override
//   List<String>? Info;
//
//   @override
//   Map<String, int>? other;
//
//   _Student({this.Name, this.Info, this.other, });
//
//   static _Student fromMap(Map<String, dynamic> map){
//     return _Student(
//       Name: map["Name"],
//       Info: map["Info"],
//       other: map["other"],
//     );
//   }
//
// }
//
// abstract class GirlStudent implements : Student, Data {
//
// /// 年龄
// Int? age;
//
// static GirlStudent create({Int? age, }){
// return _GirlStudent(age: age, );
// }
//
// static GirlStudent fromMap(Map<String, dynamic> map){
// return _GirlStudent.fromMap(map);
// }
//
// }
//
// class _GirlStudent implements GirlStudent {
//
//   @override
//   Int? age;
//
//   _GirlStudent({this.age, });
//
//   static _GirlStudent fromMap(Map<String, dynamic> map){
//     return _GirlStudent(
//       age: map["age"],
//     );
//   }
//
// }
