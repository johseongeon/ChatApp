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

  final Uri uri = Uri.parse('ws://192.168.0.12:8080/ws');

  try {
    final IOWebSocketChannel channel = IOWebSocketChannel.connect(uri);

    // initial connection
    channel.sink.add(jsonEncode({
      'username': username,
      'chat_id': chatId,
    }));

    // receive messages from the server
    channel.stream.listen((message) {
      final Map<String, dynamic> decodedMessage = jsonDecode(message);
      print('${decodedMessage['from']}: ${decodedMessage['message']}');
    }, onError: (error) {
      print('Error: $error');
      channel.sink.close();
    }, onDone: () {
      print('Disconnected from server');
    });

    // send messages to the server
    stdin.listen((List<int> data) {
      final message = utf8.decode(data).trim();
      channel.sink.add(jsonEncode({'message': message}));
      print('');
    });
  } catch (e) {
    print('Failed to connect to the server: $e');
  }
}