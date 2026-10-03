package obi

import "math"

// BucketFields are the rows' bucket fields, Bounds' order, leInf last: what
// the API sums.
func BucketFields() []string {
	out := make([]string, 0, len(Bounds)+1)
	for _, b := range Bounds {
		out = append(out, BucketField(b))
	}
	return append(out, BucketField(math.Inf(1)))
}

// Quantile reads a quantile, in milliseconds, from buckets summed across
// rows, by their fields: within the bucket it falls in, by linear
// interpolation, as Prometheus' histogram_quantile does. One past the last
// bound is that bound: how much longer, the buckets do not say. nil without a
// request.
func Quantile(buckets map[string]int64, q float64) *float64 {
	count := buckets[BucketField(math.Inf(1))]
	if count <= 0 || q < 0 || q > 1 {
		return nil
	}
	rank := q * float64(count)
	var lowerBound, lowerCount float64
	for _, bound := range Bounds {
		n := float64(buckets[BucketField(bound)])
		if n >= rank && n > lowerCount {
			v := lowerBound + (bound-lowerBound)*(rank-lowerCount)/(n-lowerCount)
			return &v
		}
		lowerBound, lowerCount = bound, math.Max(lowerCount, n)
	}
	last := Bounds[len(Bounds)-1]
	return &last
}

// AddBuckets adds one set of summed buckets to another, by field.
func AddBuckets(to, from map[string]int64) {
	for field, n := range from {
		to[field] += n
	}
}
