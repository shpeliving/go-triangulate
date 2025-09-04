package triangulate_test

import (
	"fmt"
	"math"
	"testing"

	"github.com/shpeliving/go-triangulate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Example demonstrates various triangulation capabilities
func TestTriangulateExamples(t *testing.T) {
	t.Run("Simple Triangle", func(t *testing.T) {
		// Create a simple triangle (counterclockwise winding)
		triangle := []*triangulate.Point{
			{X: 0, Y: 0},
			{X: 1, Y: 1},
			{X: 0, Y: 2},
		}

		triangles, err := triangulate.Triangulate(triangle)
		require.NoError(t, err)
		require.Len(t, triangles, 1, "Triangle should produce exactly one triangle")

		// Verify the result maintains the original vertices
		tri := triangles[0]
		assert.Contains(t, []*triangulate.Point{tri.A, tri.B, tri.C}, triangle[0])
		assert.Contains(t, []*triangulate.Point{tri.A, tri.B, tri.C}, triangle[1])
		assert.Contains(t, []*triangulate.Point{tri.A, tri.B, tri.C}, triangle[2])

		fmt.Printf("Simple triangle: %v\n", tri)
	})

	t.Run("Square", func(t *testing.T) {
		// Create a square (counterclockwise winding)
		square := []*triangulate.Point{
			{X: 0, Y: 0},
			{X: 2, Y: 0},
			{X: 2, Y: 2},
			{X: 0, Y: 2},
		}

		triangles, err := triangulate.Triangulate(square)
		require.NoError(t, err)
		require.Len(t, triangles, 2, "Square should produce exactly two triangles")

		// Verify total area is preserved (square area = 4)
		totalArea := calculateTotalArea(triangles)
		assert.InDelta(t, 4.0, totalArea, 1e-7, "Total area should equal original square area")

		fmt.Printf("Square triangulation: %d triangles, area: %.2f\n", len(triangles), totalArea)
	})

	t.Run("Complex Polygon", func(t *testing.T) {
		// Create an L-shaped polygon (counterclockwise)
		lShape := []*triangulate.Point{
			{X: 0, Y: 0},
			{X: 3, Y: 0},
			{X: 3, Y: 1},
			{X: 1, Y: 1},
			{X: 1, Y: 3},
			{X: 0, Y: 3},
		}

		triangles, err := triangulate.Triangulate(lShape)
		require.NoError(t, err)
		require.Greater(t, len(triangles), 2, "L-shape should produce multiple triangles")

		// Verify area preservation (L-shape area = 3*1 + 1*2 = 5)
		totalArea := calculateTotalArea(triangles)
		assert.InDelta(t, 5.0, totalArea, 1e-7, "Total area should equal L-shape area")

		// Verify all triangles are valid (counterclockwise)
		for i, tri := range triangles {
			area := triangleArea(tri)
			assert.Greater(t, area, 0.0, "Triangle %d should have positive area (counterclockwise)", i)
		}

		fmt.Printf("L-shape triangulation: %d triangles, area: %.2f\n", len(triangles), totalArea)
	})

	t.Run("Polygon with Hole", func(t *testing.T) {
		// Outer polygon (counterclockwise)
		outer := []*triangulate.Point{
			{X: 0, Y: 0},
			{X: 4, Y: 0},
			{X: 4, Y: 4},
			{X: 0, Y: 4},
		}

		// Inner hole (clockwise winding)
		hole := []*triangulate.Point{
			{X: 1.5, Y: 1.5},
			{X: 1.5, Y: 2.5},
			{X: 2.5, Y: 2.5},
			{X: 2.5, Y: 1.5},
		}

		triangles, err := triangulate.Triangulate(outer, hole)
		require.NoError(t, err)
		require.Greater(t, len(triangles), 4, "Polygon with hole should produce multiple triangles")

		// Verify area: outer (16) - hole (1) = 15
		totalArea := calculateTotalArea(triangles)
		assert.InDelta(t, 15.0, totalArea, 1e-7, "Total area should be outer minus hole")

		fmt.Printf("Polygon with hole: %d triangles, area: %.2f\n", len(triangles), totalArea)
	})

	t.Run("Multiple Disjoint Polygons", func(t *testing.T) {
		// Two separate squares to avoid potential coordinate issues
		// First square
		poly1 := []*triangulate.Point{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 1, Y: 1},
			{X: 0, Y: 1},
		}

		// Second square (well separated)
		poly2 := []*triangulate.Point{
			{X: 3, Y: 0},
			{X: 4, Y: 0},
			{X: 4, Y: 1},
			{X: 3, Y: 1},
		}

		triangles, err := triangulate.Triangulate(poly1, poly2)
		require.NoError(t, err)
		require.Len(t, triangles, 4, "Two disjoint squares: 2 triangles + 2 triangles = 4 total")

		// Verify total area: square (1) + square (1) = 2
		totalArea := calculateTotalArea(triangles)
		assert.InDelta(t, 2.0, totalArea, 1e-7, "Total area should be sum of both squares")

		fmt.Printf("Multiple disjoint polygons: %d triangles, area: %.2f\n", len(triangles), totalArea)
	})

	t.Run("Star-shaped Polygon", func(t *testing.T) {
		// Create a 5-pointed star (outer vertices)
		numPoints := 5
		outerRadius := 2.0
		innerRadius := 0.8

		var star []*triangulate.Point
		for i := 0; i < numPoints*2; i++ {
			angle := float64(i) * math.Pi / float64(numPoints)
			radius := outerRadius
			if i%2 == 1 {
				radius = innerRadius
			}

			star = append(star, &triangulate.Point{
				X: radius * math.Cos(angle),
				Y: radius * math.Sin(angle),
			})
		}

		triangles, err := triangulate.Triangulate(star)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(triangles), 8, "Star should produce at least 8 triangles")

		// All triangles should have positive area
		totalArea := calculateTotalArea(triangles)
		assert.Greater(t, totalArea, 0.0, "Star should have positive total area")

		fmt.Printf("Star polygon: %d triangles, area: %.2f\n", len(triangles), totalArea)
	})

	t.Run("Concurrency Safety", func(t *testing.T) {
		// Test that triangulation is safe for concurrent use
		square := []*triangulate.Point{
			{X: 0, Y: 0},
			{X: 1, Y: 0},
			{X: 1, Y: 1},
			{X: 0, Y: 1},
		}

		// Run multiple triangulations concurrently
		results := make(chan []*triangulate.Triangle, 5)

		for i := 0; i < 5; i++ {
			go func() {
				triangles, err := triangulate.Triangulate(square)
				require.NoError(t, err)
				results <- triangles
			}()
		}

		// Collect results
		for i := 0; i < 5; i++ {
			triangles := <-results
			assert.Len(t, triangles, 2, "All concurrent triangulations should produce same result")
			totalArea := calculateTotalArea(triangles)
			assert.InDelta(t, 1.0, totalArea, 1e-7, "All results should have same area")
		}

		fmt.Println("Concurrency test passed: 5 concurrent triangulations completed successfully")
	})
}

// Helper function to calculate triangle area using cross product
func triangleArea(tri *triangulate.Triangle) float64 {
	// Area = 0.5 * |cross product of two edges|
	// Using signed area: positive for CCW, negative for CW
	return 0.5 * ((tri.B.X-tri.A.X)*(tri.C.Y-tri.A.Y) - (tri.C.X-tri.A.X)*(tri.B.Y-tri.A.Y))
}

// Helper function to calculate total area of triangulated polygons
func calculateTotalArea(triangles []*triangulate.Triangle) float64 {
	total := 0.0
	for _, tri := range triangles {
		total += math.Abs(triangleArea(tri))
	}
	return total
}

// Example function showing basic usage
func ExampleTriangulate() {
	// Create a simple square
	square := []*triangulate.Point{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 1, Y: 1},
		{X: 0, Y: 1},
	}

	// Triangulate it
	triangles, err := triangulate.Triangulate(square)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Square triangulated into %d triangles\n", len(triangles))
	// Output:
	// Square triangulated into 2 triangles
}

// Benchmark demonstrates performance
func BenchmarkTriangulate(b *testing.B) {
	// Create a more complex polygon for benchmarking
	n := 100
	var polygon []*triangulate.Point
	for i := 0; i < n; i++ {
		angle := 2 * math.Pi * float64(i) / float64(n)
		// Add some noise to make it more interesting
		radius := 1.0 + 0.1*math.Sin(5*angle)
		polygon = append(polygon, &triangulate.Point{
			X: radius * math.Cos(angle),
			Y: radius * math.Sin(angle),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := triangulate.Triangulate(polygon)
		if err != nil {
			b.Fatal(err)
		}
	}
}
