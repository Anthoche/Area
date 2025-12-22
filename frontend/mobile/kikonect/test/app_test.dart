import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:kikonect/screens/home_page.dart';
import 'package:kikonect/widgets/app_text_field.dart';
import 'package:kikonect/widgets/primary_button.dart';
import 'package:kikonect/widgets/search_bar.dart';
import 'package:kikonect/widgets/service_card.dart';

// Helper to find the TextField inside our custom AppTextField widget
Finder findAppTextField(String label) {
  return find.descendant(
    of: find.widgetWithText(AppTextField, label),
    matching: find.byType(TextField),
  );
}

void main() {
  setUpAll(() {
    dotenv.testLoad(fileInput: 'API_URL=http://test.com');
  });

  group('Shared Widgets Tests', () {
    testWidgets('AppTextField renders label and respects obscureText', (
        WidgetTester tester) async {
      await tester.pumpWidget(
        const MaterialApp(
          home: Scaffold(
            body: AppTextField(label: 'Test Label', obscure: true),
          ),
        ),
      );

      expect(find.text('Test Label'), findsOneWidget);
      final textField = tester.widget<TextField>(find.byType(TextField));
      expect(textField.obscureText, isTrue);
    });

    testWidgets('PrimaryButton renders text and triggers callback', (
        WidgetTester tester) async {
      bool pressed = false;
      await tester.pumpWidget(
        MaterialApp(
          home: Scaffold(
            body: PrimaryButton(
              text: 'Click Me',
              onPressed: () => pressed = true,
            ),
          ),
        ),
      );

      expect(find.text('Click Me'), findsOneWidget);
      await tester.tap(find.byType(PrimaryButton));
      expect(pressed, isTrue);
    });
  });

  group('Homepage Tests', () {
    testWidgets('renders correctly', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: Homepage()));

      expect(find.text('My Konect'), findsOneWidget);
      expect(find.byType(AppSearchBar), findsOneWidget);
      expect(find.text('My Konects'), findsOneWidget);

      // Check for FilterTags (by text, assuming 'test 1' is hardcoded in homepage)
      expect(find.text('test 1'), findsOneWidget);

      // Check for ServiceCards
      expect(find.text('Push To Ping'), findsOneWidget);
      expect(find.byType(ServiceCard), findsAtLeastNWidgets(1));
    });

    testWidgets('opens drawer', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: Homepage()));

      await tester.tap(find.byTooltip('Open navigation menu'));
      await tester.pumpAndSettle();

      expect(find.text('Menu'), findsOneWidget);
      expect(find.text('test'), findsOneWidget);
    });

    testWidgets('FAB shows Create Area page', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: Homepage()));

      await tester.tap(find.byIcon(Icons.add));
      await tester.pumpAndSettle();
      
      expect(find.text('Create'), findsOneWidget);
    });
  });
}
