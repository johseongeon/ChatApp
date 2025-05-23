import 'package:flutter/material.dart';

void main() {
  runApp(const MyApp());
}
class MyApp extends StatelessWidget {
 const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        bottomNavigationBar: BottomNavigationBar(
          items: [
            BottomNavigationBarItem(
              icon: Icon(Icons.group),
              label: 'friends',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.message),
              label: 'message',
            ),
            BottomNavigationBarItem(
              icon: Icon(Icons.settings),
              label: 'settings')
          ]
          )
      )
    );
  }
}