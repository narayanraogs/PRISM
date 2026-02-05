import 'dart:math';

class DataPoint {
  final double x;
  final double y;
  DataPoint(this.x, this.y);
}

/// LTTB (Largest Triangle Three Buckets) Downsampling Algorithm
/// 
/// This algorithm reduces the number of points in a time series while
/// preserving its visual characteristics (peaks and troughs).
/// 
/// [data] - The raw list of DataPoints.
/// [threshold] - The number of points to reduce to.
List<DataPoint> lttb(List<DataPoint> data, int threshold) {
  int dataLength = data.length;
  if (threshold >= dataLength || threshold == 0) return data;

  List<DataPoint> sampled = [];
  
  // Bucket size. Leave room for start and end points
  double every = (dataLength - 2) / (threshold - 2);

  int a = 0; // The last selected point index
  sampled.add(data[0]); // Always add the first point

  for (int i = 0; i < threshold - 2; i++) {
    // Calculate point average for next bucket (used as a reference point 'C')
    double avgX = 0;
    double avgY = 0;
    int avgRangeStart = ((i + 1) * every).floor() + 1;
    int avgRangeEnd = ((i + 2) * every).floor() + 1;
    avgRangeEnd = avgRangeEnd < dataLength ? avgRangeEnd : dataLength;

    int avgRangeLength = avgRangeEnd - avgRangeStart;

    for (int j = avgRangeStart; j < avgRangeEnd; j++) {
      avgX += data[j].x;
      avgY += data[j].y;
    }
    avgX /= avgRangeLength;
    avgY /= avgRangeLength;

    // Get the range for this bucket (current points to pick from)
    int rangeOffs = (i * every).floor() + 1;
    int rangeTo = ((i + 1) * every).floor() + 1;

    // Current fixed point 'A'
    double pointAx = data[a].x;
    double pointAy = data[a].y;

    double maxArea = -1;
    int nextAIndex = rangeOffs;

    for (int j = rangeOffs; j < rangeTo; j++) {
      // Calculate triangle area formed by point A, candidate point from current bucket, and average of next bucket
      // Area = |(Ax(By - Cy) + Bx(Cy - Ay) + Cx(Ay - By)) / 2|
      double area = ((pointAx - avgX) * (data[j].y - pointAy) -
              (pointAx - data[j].x) * (avgY - pointAy))
          .abs() * 0.5;
      
      if (area > maxArea) {
        maxArea = area;
        nextAIndex = j;
      }
    }

    sampled.add(data[nextAIndex]); 
    a = nextAIndex; // This becomes the starting point for the next bucket
  }

  sampled.add(data[dataLength - 1]); // Always add the last point
  return sampled;
}
