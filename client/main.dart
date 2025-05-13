import 'dart:convert';
import 'dart:io';

import 'package:web_socket_channel/io.dart';

void main(List<String> args) async {
  if (args.length < 2) {
    print('Usage: dart run client.dart <username> <chat_id>');
    return;
  }

  final String username = args[0];
  final String chatId = args[1];

  final Uri uri = Uri.parse('ws://localhost:8080/ws');

  try {
    final IOWebSocketChannel channel = IOWebSocketChannel.connect(uri);

    // 서버에 사용자 정보 전송
    channel.sink.add(jsonEncode({
      'username': username,
      'chat_id': chatId,
    }));

    // 서버에서 메시지 수신
    channel.stream.listen((message) {
      final Map<String, dynamic> decodedMessage = jsonDecode(message);
      print('${decodedMessage['from']}: ${decodedMessage['message']}');
    }, onError: (error) {
      print('Error: $error');
      channel.sink.close();
    }, onDone: () {
      print('Disconnected from server');
    });

    // 사용자로부터 입력 받아서 서버에 전송
    stdin.listen((List<int> data) {
      final message = utf8.decode(data).trim();
      channel.sink.add(jsonEncode({'message': message}));
      print(''); // 개행
    });
  } catch (e) {
    print('Failed to connect to the server: $e');
  }
}