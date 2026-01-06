// This is a basic Flutter widget test.
//
// To perform an interaction with a widget in your test, use the WidgetTester
// utility in the flutter_test package. For example, you can send tap and scroll
// gestures. You can also use WidgetTester to find child widgets in the widget
// tree, read text, and verify that the values of widget properties are correct.

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:kikonect/widgets/service_card.dart';

void main() {
  testWidgets('ServiceCard renders and responds to taps',
      (WidgetTester tester) async {
    var tapped = false;
    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: ServiceCard(
            title: 'Test Service',
            color: Colors.blue,
            onTap: () => tapped = true,
          ),
        ),
      ),
    );

    expect(find.text('Test Service'), findsOneWidget);
    await tester.tap(find.byType(ServiceCard));
    expect(tapped, isTrue);
  });
}
