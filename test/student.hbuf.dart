
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

