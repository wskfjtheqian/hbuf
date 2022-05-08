import 'dart:async';
import 'dart:io';

import 'package:cookie_jar/cookie_jar.dart';
import 'package:hbuf_dart/hbuf/server.dart';
import 'package:hbuf_dart/http/client.dart';
import 'package:hbuf_dart/http/server.dart';
import 'package:test/test.dart';

import '../bin/people.hbuf.dart';


class StudentServerImp extends StudentServer{
  @override
  Future<GetAddressRes> getAddress(GetAddressReq userId, [Context? ctx]) async {
    return GetAddressRes(address: '了很多年吧是');
  }

  @override
  Future<People> getNumber(People userId, [Context? ctx]) {
    // TODO: implement getNumber
    throw UnimplementedError();
  }

  @override
  Future<People> getPeople(People userId, [Context? ctx]) {
    // TODO: implement getPeople
    throw UnimplementedError();
  }

}


void main() {
  group('hbuf http tests', () {
    test('Http Client', () async {
      var cookie = CookieJar();
      var client = HttpClientJson(
        baseUrl: "http://localhost:8080",
      );
      client.insertRequestInterceptor((request, data, next) async {
        request.cookies.addAll(await cookie.loadForRequest(request.uri));
        next?.invoke!(request, data, next.next);
      });
      client.insertResponseInterceptor((request, response, data, next) async {
        await cookie.saveFromResponse(request.uri, response.cookies);
        return await next?.invoke!(request, response, data, next.next) ?? data;
      });

      var people = StudentServerClient(client);
      try {
        var name = await people.getAddress(GetAddressReq());
        print(name.address + '\n');
      } catch (e) {
        print(e);
      }
    });

    test("Http Server", () async {
      var server = await HttpServer.bind("0.0.0.0", 8080);

      var router = HttpServerJson();
      router.add(StudentServerRouter(StudentServerImp()));

      server.listen(router.onData);
      await Completer.sync().future;
    }, timeout: Timeout(Duration(days: 100)));
  });
}
