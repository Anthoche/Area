import 'dart:async';

import 'package:app_links/app_links.dart';
import 'package:flutter/material.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import '../services/oauth_service.dart';
import '../utils/ui_feedback.dart';
import 'login_page.dart';

/// Displays the profile screen and connected services.
class ProfilePage extends StatefulWidget {
  const ProfilePage({super.key});

  @override
  State<ProfilePage> createState() => _ProfilePageState();
}

class _ProfilePageState extends State<ProfilePage> {
  final _storage = const FlutterSecureStorage();
  final _oauthService = OAuthService();
  final _appLinks = AppLinks();
  StreamSubscription? _sub;
  final Set<String> _busy = {};
  bool _loading = true;
  Map<String, String?> _tokens = {};

  @override
  void initState() {
    super.initState();
    _loadTokens();
    _initDeepLinks();
  }

  @override
  void dispose() {
    _sub?.cancel();
    super.dispose();
  }

  Future<void> _loadTokens() async {
    final googleToken = await _storage.read(key: 'google_token_id');
    final githubToken = await _storage.read(key: 'github_token_id');
    if (mounted) {
      setState(() {
        _tokens = {
          'google_token_id': googleToken,
          'github_token_id': githubToken,
        };
        _loading = false;
      });
    }
  }

  Future<void> _initDeepLinks() async {
    _sub = _appLinks.uriLinkStream.listen((uri) {
      _handleDeepLink(uri);
    });
  }

  Future<void> _handleDeepLink(Uri uri) async {
    final error = uri.queryParameters['error'];
    if (error != null) {
      showAppSnackBar(context, "OAuth error: $error", isError: true);
      return;
    }

    final tokenId = uri.queryParameters['token_id'];
    final googleEmail = uri.queryParameters['google_email'];
    final githubEmail = uri.queryParameters['github_email'];
    final githubLogin = uri.queryParameters['github_login'];
    if (tokenId != null && tokenId.isNotEmpty) {
      if (githubLogin != null || githubEmail != null) {
        await _storage.write(key: 'github_token_id', value: tokenId);
      } else if (googleEmail != null) {
        await _storage.write(key: 'google_token_id', value: tokenId);
      } else {
        await _storage.write(key: 'github_token_id', value: tokenId);
      }
      await _loadTokens();
      return;
    }

    final code = uri.queryParameters['code'];
    final state = uri.queryParameters['state'];
    if (code != null && code.isNotEmpty) {
      try {
        final result =
            await _oauthService.exchangeCodeForToken(code, state: state);
        final token = result['token_id'] ?? result['token'];
        if (token != null) {
          await _storage.write(key: 'google_token_id', value: token.toString());
        }
        final userId = result['user_id'];
        if (userId != null) {
          await _storage.write(key: 'user_id', value: userId.toString());
        }
        await _loadTokens();
      } catch (e) {
        showAppSnackBar(context, "OAuth error: $e", isError: true);
      }
    }
  }

  Future<void> _connectService(_ServiceEntry entry) async {
    if (_busy.contains(entry.id)) return;
    setState(() => _busy.add(entry.id));
    try {
      await _oauthService.signInWith(entry.id);
    } catch (e) {
      showAppSnackBar(context, "Connection failed: $e", isError: true);
    } finally {
      if (mounted) {
        setState(() => _busy.remove(entry.id));
      }
    }
  }

  Future<void> _disconnectService(_ServiceEntry entry) async {
    if (_busy.contains(entry.id)) return;
    setState(() => _busy.add(entry.id));
    try {
      await _storage.delete(key: entry.tokenKey);
      await _loadTokens();
    } finally {
      if (mounted) {
        setState(() => _busy.remove(entry.id));
      }
    }
  }

  Future<void> _logout() async {
    await _storage.deleteAll();
    if (mounted) {
      Navigator.pushAndRemoveUntil(
        context,
        MaterialPageRoute(builder: (_) => const LoginPage()),
        (route) => false,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.white,
      appBar: AppBar(
        backgroundColor: Colors.white,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Colors.black),
          onPressed: () => Navigator.pop(context),
        ),
        title: const Text(
          "Profile",
          style: TextStyle(
            color: Colors.black,
            fontWeight: FontWeight.bold,
          ),
        ),
        centerTitle: true,
      ),
      body: _loading
          ? const Center(child: CircularProgressIndicator())
          : ListView(
              padding: const EdgeInsets.all(20),
              children: [
                const Text(
                  "Connected services",
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.w600,
                    color: Colors.black87,
                  ),
                ),
                const SizedBox(height: 16),
                ..._ServiceEntry.entries.map((entry) {
                  final token = _tokens[entry.tokenKey];
                  final connected = token != null && token.isNotEmpty;
                  final busy = _busy.contains(entry.id);
                  return Container(
                    margin: const EdgeInsets.only(bottom: 12),
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Colors.grey[100],
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(color: Colors.grey[300]!),
                    ),
                    child: Row(
                      children: [
                        entry.iconPath != null
                            ? Image.asset(entry.iconPath!,
                                height: 36, width: 36)
                            : const Icon(Icons.apps, size: 36),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                entry.name,
                                style: const TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                              const SizedBox(height: 4),
                              Text(
                                entry.description,
                                style: TextStyle(color: Colors.grey[600]),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(width: 12),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor:
                                connected ? Colors.red : Colors.black,
                            foregroundColor: Colors.white,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(20),
                            ),
                          ),
                          onPressed: busy
                              ? null
                              : connected
                                  ? () => _disconnectService(entry)
                                  : () => _connectService(entry),
                          child: Text(connected ? "Disconnect" : "Connect"),
                        ),
                      ],
                    ),
                  );
                }),
                const SizedBox(height: 12),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.black,
                    foregroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(30),
                    ),
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                  onPressed: _logout,
                  child: const Text("Logout"),
                ),
              ],
            ),
    );
  }
}

class _ServiceEntry {
  final String id;
  final String name;
  final String description;
  final String tokenKey;
  final String? iconPath;

  const _ServiceEntry({
    required this.id,
    required this.name,
    required this.description,
    required this.tokenKey,
    this.iconPath,
  });

  static const entries = [
    _ServiceEntry(
      id: 'google',
      name: 'Google',
      description: 'Gmail, Calendar',
      tokenKey: 'google_token_id',
      iconPath: 'lib/assets/G_logo.png',
    ),
    _ServiceEntry(
      id: 'github',
      name: 'GitHub',
      description: 'Issues, Pull Requests',
      tokenKey: 'github_token_id',
      iconPath: 'lib/assets/github_logo.png',
    ),
  ];
}
