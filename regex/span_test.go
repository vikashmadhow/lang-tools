package regex

import (
	"fmt"
	"slices"
	"testing"
)

func TestSpanIntersect(t *testing.T) {
	s1 := span{0, 10}
	s2 := span{5, 15}
	s3 := span{10, 20}
	s4 := span{15, 25}
	s5 := span{20, 30}

	if !s1.intersect(s2) {
		t.Error("{}, {} should intersect", s1, s2)
	}
	if !s2.intersect(s1) {
		t.Error("{}, {} should intersect", s2, s1)
	}
	if !s1.intersect(s3) {
		t.Error("{}, {} should intersect", s1, s3)
	}
	if !s3.intersect(s1) {
		t.Error("{}, {} should intersect", s3, s1)
	}
	if s1.intersect(s4) {
		t.Error("{}, {} should not intersect", s1, s4)
	}
	if s4.intersect(s1) {
		t.Error("{}, {} should not intersect", s4, s1)
	}
	if s1.intersect(s5) {
		t.Error("{}, {} should not intersect", s1, s5)
	}
	if s5.intersect(s1) {
		t.Error("{}, {} should not intersect", s5, s1)
	}
}

func TestCompact(t *testing.T) {
	s := spanSet{
		{0, 10},
		{3, 5},
		{4, 12},
		{15, 30},
		{14, 21},
		{27, 29},
		{28, 33},
	}
	actual := s.compact()
	expected := spanSet{
		{0, 12},
		{14, 33},
	}
	if !slices.Equal(actual, expected) {
		fmt.Println(expected)
		fmt.Println(actual)
		t.Error("expected {}, actual {}", expected, actual)
	}
}

func TestMinus(t *testing.T) {
	s1 := spanSet{
		{0, 10},
		{3, 5},
		{4, 12},
		{16, 30},
		{15, 21},
		{27, 29},
		{28, 33},
	}
	s2 := spanSet{
		{2, 3},
		{5, 7},
		{14, 18},
		{21, 35},
	}
	expected := spanSet{
		{0, 1},
		{4, 4},
		{8, 12},
		{19, 20},
	}
	actual := s1.minus(s2)
	if !slices.Equal(actual, expected) {
		fmt.Println(expected)
		fmt.Println(actual)
		t.Error("expected", expected, "actual", actual)
	}
}

func TestMinus2(t *testing.T) {
	s1 := spanSet{
		{5, 10},
		{3, 5},
		{4, 12},
		{16, 30},
		{15, 21},
		{27, 29},
		{28, 33},
	}
	//expected := spanSet{
	//	{0, 1},
	//	{4, 4},
	//	{8, 12},
	//	{19, 20},
	//}
	s2 := allUnicode.minus(s1)
	s4 := s1.compact()
	fmt.Println(s4)
	fmt.Println(s2)
	s3 := allUnicode.minus(s2)

	if !slices.Equal(s3, s4) {
		t.Error("expected", s4, "actual", s3)
	}
}
