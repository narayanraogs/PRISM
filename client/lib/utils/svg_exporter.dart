import '../screens/path_loss_planner_screen.dart';

class SvgExporter {
  static const double nw = 180;
  static const double nh = 90;
  static const double gx = 50;
  static const double gy = 80;

  static String generate(PlannerNode hub) {
    final Map<PlannerNode, double> wMap = {};

    double calcW(PlannerNode n) {
      if (n.children.isEmpty) {
        wMap[n] = nw;
        return nw;
      }
      double sum = 0;
      for (var c in n.children) {
        sum += calcW(c);
      }
      sum += (n.children.length - 1) * gx;
      final w = sum > nw ? sum : nw;
      wMap[n] = w;
      return w;
    }

    final topC = hub.children
        .where((c) => c.direction == NodeDirection.up)
        .toList();
    final botC = hub.children
        .where((c) => c.direction == NodeDirection.down)
        .toList();

    double topW = 0;
    if (topC.isNotEmpty) {
      for (var c in topC) {
        topW += calcW(c);
      }
      topW += (topC.length - 1) * gx;
    }

    double botW = 0;
    if (botC.isNotEmpty) {
      for (var c in botC) {
        botW += calcW(c);
      }
      botW += (botC.length - 1) * gx;
    }

    final hubW = nw;
    double maxW = topW;
    if (botW > maxW) maxW = botW;
    if (hubW > maxW) maxW = hubW;

    int getDepth(PlannerNode n) {
      if (n.children.isEmpty) return 1;
      int md = 0;
      for (var c in n.children) {
        final d = getDepth(c);
        if (d > md) md = d;
      }
      return md + 1;
    }

    int topDepth = 0;
    for (var c in topC) {
      final d = getDepth(c);
      if (d > topDepth) topDepth = d;
    }

    int botDepth = 0;
    for (var c in botC) {
      final d = getDepth(c);
      if (d > botDepth) botDepth = d;
    }

    final double contentW = maxW + 200;
    final double topH = topDepth * (nh + gy);
    final double botH = botDepth * (nh + gy);
    final double contentH = topH + nh + botH + 200;

    // Fix aspect ratio for Landscape A3 (420mm x 297mm) -> ~1.414
    const double a3Ratio = 420 / 297;
    double totalW = contentW;
    double totalH = contentH;

    if (totalW / totalH > a3Ratio) {
      // Content is wider than A3 ratio, expand height to fit ratio
      totalH = totalW / a3Ratio;
    } else {
      // Content is taller than A3 ratio, expand width to fit ratio
      totalW = totalH * a3Ratio;
    }

    final double cx = totalW / 2;
    // Center the entire visual structure vertically
    // The visual structure height is (topH + nh + botH)
    // We center the hub in the vertical middle of the final canvas
    final double cy = totalH / 2; 

    List<String> elements = [];

    void line(double x1, double y1, double x2, double y2) {
      elements.add(
        '<line x1="${x1.toStringAsFixed(1)}" y1="${y1.toStringAsFixed(1)}" x2="${x2.toStringAsFixed(1)}" y2="${y2.toStringAsFixed(1)}" stroke="#cbd5e1" stroke-width="2"/>',
      );
    }

    void box(double bx, double by, PlannerNode n) {
      final x = bx - nw / 2;
      final y = by - nh / 2;
      String color = "#ffffff";
      if (n.type == NodeType.hub) {
        color = "#f3e8ff";
      } else if (n.type == NodeType.branching) {
        color = "#ffedd5";
      } else if (n.type == NodeType.source) {
        color = "#e0f2fe";
      } else if (n.type == NodeType.instrument) {
        color = "#dcfce7";
      } else if (n.type == NodeType.converter) {
        color = "#fce7f3";
      }

      String border =
          (n.type == NodeType.hub ||
              n.type == NodeType.source ||
              n.type == NodeType.instrument)
          ? "#4f46e5"
          : "#94a3b8";

      elements.add(
        '<rect x="${x.toStringAsFixed(1)}" y="${y.toStringAsFixed(1)}" width="$nw" height="$nh" rx="8" fill="$color" stroke="$border" stroke-width="2"/>',
      );

      // Label wrapping logic
      final label = n.label;
      final words = label.split(' ');
      final List<String> lines = [];
      String currentLine = "";
      for (var word in words) {
        if (("$currentLine $word").length < 24) {
          currentLine += (currentLine.isEmpty ? "" : " ") + word;
        } else {
          if (currentLine.isNotEmpty) lines.add(currentLine);
          currentLine = word;
        }
      }
      if (currentLine.isNotEmpty) lines.add(currentLine);

      // Cap at 2 lines for label to avoid layout break
      if (lines.length > 2) {
        lines[1] = "${lines[1]}...";
        lines.removeRange(2, lines.length);
      }

      double textY = y + (lines.length == 1 ? 25 : 20);
      for (var line in lines) {
        elements.add(
          '<text x="${bx.toStringAsFixed(1)}" y="${textY.toStringAsFixed(1)}" font-family="sans-serif" font-size="12" font-weight="bold" text-anchor="middle" fill="#1e293b">$line</text>',
        );
        textY += 14;
      }

      elements.add(
        '<text x="${bx.toStringAsFixed(1)}" y="${(y + 55).toStringAsFixed(1)}" font-family="sans-serif" font-size="10" text-anchor="middle" fill="#64748b">${n.type.name.toUpperCase()}</text>',
      );

      String sub;
      if (n.type == NodeType.converter) {
        sub = "LO: ${n.loOffset} MHz";
      } else if (n.type == NodeType.source) {
        sub = "${n.sourceFrequency} MHz | ${n.lossDb} dBm";
      } else if (n.calibratedCableId != null) {
        sub = "${n.calibratedCableId}";
      } else {
        sub = "${n.lossDb} dB";
      }

      elements.add(
        '<text x="${bx.toStringAsFixed(1)}" y="${(y + 75).toStringAsFixed(1)}" font-family="sans-serif" font-size="11" font-weight="bold" text-anchor="middle" fill="#3b82f6">$sub</text>',
      );
    }

    void layoutTop(
      List<PlannerNode> nodes,
      double px,
      double py,
      double availableW,
    ) {
      if (nodes.isEmpty) return;
      double startX = px - availableW / 2;
      for (var c in nodes) {
        double cw = wMap[c]!;
        double childCx = startX + cw / 2;
        double childCy = py - nh - gy;

        line(childCx, childCy + nh / 2, childCx, py - gy / 2);
        box(childCx, childCy, c);
        layoutTop(c.children, childCx, childCy, cw);

        startX += cw + gx;
      }

      if (nodes.length > 1) {
        double firstCx = px - availableW / 2 + wMap[nodes.first]! / 2;
        double lastCx = px + availableW / 2 - wMap[nodes.last]! / 2;
        line(firstCx, py - gy / 2, lastCx, py - gy / 2);
      }
      if (nodes.isNotEmpty) {
        line(px, py - gy / 2, px, py - nh / 2);
      }
    }

    void layoutBot(
      List<PlannerNode> nodes,
      double px,
      double py,
      double availableW,
    ) {
      if (nodes.isEmpty) return;
      double startX = px - availableW / 2;
      for (var c in nodes) {
        double cw = wMap[c]!;
        double childCx = startX + cw / 2;
        double childCy = py + nh + gy;

        line(childCx, childCy - nh / 2, childCx, py + gy / 2);
        box(childCx, childCy, c);
        layoutBot(c.children, childCx, childCy, cw);

        startX += cw + gx;
      }

      if (nodes.length > 1) {
        double firstCx = px - availableW / 2 + wMap[nodes.first]! / 2;
        double lastCx = px + availableW / 2 - wMap[nodes.last]! / 2;
        line(firstCx, py + gy / 2, lastCx, py + gy / 2);
      }
      if (nodes.isNotEmpty) {
        line(px, py + gy / 2, px, py + nh / 2);
      }
    }

    box(cx, cy, hub);
    if (topC.isNotEmpty) {
      layoutTop(topC, cx, cy, topW);
    }
    if (botC.isNotEmpty) {
      layoutBot(botC, cx, cy, botW);
    }

    final String svgStr =
        '''<?xml version="1.0" standalone="no"?>
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" 
  "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 $totalW $totalH" width="$totalW" height="$totalH" style="background-color: #f8fafc;">
${elements.join('\\n')}
</svg>''';

    return svgStr;
  }
}
