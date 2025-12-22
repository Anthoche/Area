import 'package:flutter/material.dart';
import '../widgets/filter_tag.dart';
import '../widgets/service_card.dart';
import '../widgets/search_bar.dart';
import 'create_area_page.dart';

/// Home screen showing saved Konects and quick actions.
class Homepage extends StatelessWidget {
  const Homepage({super.key});

  @override
  Widget build(BuildContext context) {
    final List<Map<String, dynamic>> items = List.generate(8, (index) {
      return {
        'title': index == 0 ? 'Push To Ping' : 'Service ${index + 1}',
        'color': _getColor(index),
        'icons': index == 0 ? [Icons.code, Icons.message] : <IconData>[],
      };
    });

    return Scaffold(
      backgroundColor: Colors.white,
      drawer: Drawer(
        backgroundColor: Colors.white,
        child: ListView(
          padding: EdgeInsets.zero,
          children: [
            const DrawerHeader(
              decoration: BoxDecoration(color: Colors.white),
              child: Center(
                child: Text(
                  'Menu',
                  style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            ListTile(
              title: const Text('test'),
              onTap: () {},
            ),
          ],
        ),
      ),
      appBar: AppBar(
        backgroundColor: Colors.white,
        surfaceTintColor: Colors.transparent,
        elevation: 0,
        title: const Text(
          'My Konect',
          style: TextStyle(
            color: Colors.black,
            fontWeight: FontWeight.bold,
            fontFamily: 'Serif',
            fontSize: 22,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.person_outline, color: Colors.black),
            onPressed: () {},
          ),
          const SizedBox(width: 8),
        ],
      ),
      body: CustomScrollView(
        slivers: [
          const SliverToBoxAdapter(
            child: Padding(
              padding: EdgeInsets.all(16.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Search_bar(),
                  SizedBox(height: 20),
                  SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: Row(
                      children: [
                        FilterTag(label: 'test 1'),
                        FilterTag(label: 'test 2'),
                        FilterTag(label: 'test 3'),
                        FilterTag(label: 'test 4'),
                        FilterTag(label: 'test 5'),
                      ],
                    ),
                  ),
                  SizedBox(height: 24),
                  Text(
                    'My Konects',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.w500,
                      color: Colors.black54,
                    ),
                  ),
                  SizedBox(height: 16),
                ],
              ),
            ),
          ),
          SliverPadding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            sliver: SliverGrid(
              gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                crossAxisCount: 2,
                mainAxisSpacing: 16,
                crossAxisSpacing: 16,
                childAspectRatio: 1.3,
              ),
              delegate: SliverChildBuilderDelegate(
                    (context, index) {
                  final item = items[index];
                  return ServiceCard(
                    title: item['title'],
                    color: item['color'],
                  );
                },
                childCount: items.length,
              ),
            ),
          ),
          const SliverToBoxAdapter(child: SizedBox(height: 100)),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        backgroundColor: const Color(0xFF7209B7),
        foregroundColor: Colors.white,
        shape: const CircleBorder(),
        elevation: 4,
        onPressed: () {
          Navigator.push(
            context,
            MaterialPageRoute(builder: (context) => const CreateAreaPage()),
          );
        },
        child: const Icon(Icons.add, size: 30),
      ),
    );
  }

  /// Opens a contextual menu anchored to the floating action button.
  void _showAddOptions(BuildContext context) {
    final RenderBox overlay = Overlay
        .of(context)
        .context
        .findRenderObject() as RenderBox;

    showMenu(
      context: context,
      position: RelativeRect.fromLTRB(
        overlay.size.width - 150,
        overlay.size.height - 200,
        16,
        0,
      ),
      items: [
        const PopupMenuItem<String>(
          value: 'option1',
          child: Row(
            children: [
              Text('Option 1'),
            ],
          ),
        ),
      ],
      elevation: 8,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
    ).then((value) {
      if (value == 'option1') {}
    });
  }

  /// Rotates through a small palette to colorize service cards.
  Color _getColor(int index) {
    final colors = [
      const Color(0xFF00D2FF),
      const Color(0xFFFF4081),
      const Color(0xFFFF4081),
      const Color(0xFF00E676),
      const Color(0xFFD500F9),
    ];
    return colors[index % colors.length];
  }
}
