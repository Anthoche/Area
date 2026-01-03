import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:kikonect/screens/home_page.dart';
import 'package:kikonect/services/api_service.dart';
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

class FakeApiService extends ApiService {
  FakeApiService(this.workflows);

  final List<dynamic> workflows;

  @override
  Future<List<dynamic>> getWorkflows() async {
    return workflows;
  }
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
      final workflows = [
        {
          'id': 1,
          'name': 'Push To Ping',
          'trigger_type': 'manual',
          'enabled': true,
          'trigger_config': '{}',
        },
      ];

      await tester.pumpWidget(
        MaterialApp(
          home: Homepage(apiService: FakeApiService(workflows)),
        ),
      );
      await tester.pumpAndSettle();

      expect(find.text('My Konect'), findsOneWidget);
      expect(find.text('My Konects'), findsOneWidget);
      expect(find.byType(AppSearchBar), findsOneWidget);
      expect(find.text('Push To Ping'), findsOneWidget);
      expect(find.text('MANUAL'), findsOneWidget);
      expect(find.byType(ServiceCard), findsAtLeastNWidgets(1));
    });

    testWidgets('FAB shows Create Area page', (WidgetTester tester) async {
      await tester.pumpWidget(
        MaterialApp(
          home: Homepage(apiService: FakeApiService([])),
        ),
      );
      await tester.pumpAndSettle();

      await tester.tap(find.byIcon(Icons.add));
      await tester.pumpAndSettle();
      
      expect(find.text('Create'), findsOneWidget);
    });
  });
}
