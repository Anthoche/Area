import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';
import 'package:poc_area/screens/login_page.dart';
import 'package:poc_area/screens/register_middle_page.dart';
import 'package:poc_area/screens/register_page.dart';
import 'package:poc_area/screens/home_page.dart';
import 'package:poc_area/widgets/app_text_field.dart';
import 'package:poc_area/widgets/primary_button.dart';
import 'package:poc_area/widgets/search_bar.dart';
import 'package:poc_area/widgets/service_card.dart';

// Helper to find the TextField inside our custom AppTextField widget
Finder findAppTextField(String label) {
  return find.descendant(
    of: find.widgetWithText(AppTextField, label),
    matching: find.byType(TextField),
  );
}

void main() {
  setUpAll(() async {
    // Setup mock environment variables to prevent errors with dotenv
    await dotenv.testLoad(fileInput: 'API_URL=http://test.com');
  });

  group('Shared Widgets Tests', () {
    testWidgets('AppTextField renders label and respects obscureText', (WidgetTester tester) async {
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

    testWidgets('PrimaryButton renders text and triggers callback', (WidgetTester tester) async {
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

  group('LoginPage Tests', () {
    testWidgets('renders correctly', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: LoginPage()));
      expect(find.text('Email'), findsOneWidget);
      expect(find.text('Password'), findsOneWidget);
      expect(find.text('Login'), findsOneWidget);
    });

    testWidgets('shows error on empty fields', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: LoginPage()));
      await tester.tap(find.text('Login'));
      await tester.pump();
      expect(find.text('Please fill in all fields.'), findsOneWidget);
      await tester.tap(find.text('OK'));
      await tester.pump();
    });

    testWidgets('shows error on invalid email', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: LoginPage()));
      await tester.enterText(findAppTextField('Email'), 'invalid');
      await tester.enterText(findAppTextField('Password'), '123');
      await tester.tap(find.text('Login'));
      await tester.pump();
      expect(find.text('Please enter a valid email address.'), findsOneWidget);
      await tester.tap(find.text('OK'));
      await tester.pump();
    });
    
    testWidgets('navigates to RegisterMiddlePage', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: LoginPage()));
      await tester.tap(find.text('Sign up'));
      await tester.pumpAndSettle();
      expect(find.byType(RegisterMiddlePage), findsOneWidget);
    });
  });

  group('RegisterMiddlePage Tests', () {
    testWidgets('renders correctly', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterMiddlePage()));
      expect(find.text('Create an account'), findsOneWidget);
      expect(find.text('Enter your email to sign up for this app'), findsOneWidget);
    });

    testWidgets('navigates to RegisterPage', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterMiddlePage()));
      await tester.enterText(findAppTextField('Email'), 'test@test.com');
      await tester.tap(find.text('Continue'));
      await tester.pumpAndSettle();
      expect(find.byType(RegisterPage), findsOneWidget);
      expect(find.text('test@test.com'), findsOneWidget);
    });
  });

  group('RegisterPage Tests', () {
    testWidgets('renders correctly', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterPage()));
      expect(find.text('First Name'), findsOneWidget);
      expect(find.text('Register'), findsOneWidget);
    });

    testWidgets('validates empty fields', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterPage()));
      await tester.tap(find.text('Register'));
      await tester.pump();
      expect(find.text('Please fill in all fields.'), findsOneWidget);
      await tester.tap(find.text('OK'));
      await tester.pump();
    });

    testWidgets('validates passwords match', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: RegisterPage()));
      // Fill all but make passwords mismatch
      await tester.enterText(findAppTextField('First Name'), 'A');
      await tester.enterText(findAppTextField('Last Name'), 'B');
      await tester.enterText(findAppTextField('Email'), 'a@b.c');
      await tester.enterText(findAppTextField('Password'), '123');
      await tester.enterText(findAppTextField('Confirm Password'), '456');
      
      await tester.tap(find.text('Register'));
      await tester.pump();
      
      expect(find.text('Passwords do not match'), findsOneWidget);
      await tester.tap(find.text('OK'));
      await tester.pump();
    });
  });

  group('Homepage Tests', () {
    testWidgets('renders correctly', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: Homepage()));
      
      expect(find.text('My Konect'), findsOneWidget);
      expect(find.byType(Search_bar), findsOneWidget);
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

    testWidgets('FAB shows popup menu', (WidgetTester tester) async {
      await tester.pumpWidget(const MaterialApp(home: Homepage()));
      
      await tester.tap(find.byIcon(Icons.add));
      await tester.pumpAndSettle();
      
      expect(find.text('Option 1'), findsOneWidget);
    });
  });
}
