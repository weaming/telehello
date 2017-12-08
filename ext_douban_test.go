package main

import (
	"fmt"
	"testing"
)

func TestDouban(t *testing.T) {
	city := "深圳"
	hotMovies, err := GetMovieInTheaters(city)
	if err != nil {
		t.Error(err)
	}
	//fmt.Printf("%#v\n", mvList)
	mvs := hotMovies.FilterSubject(func(i int, mv MovieSubject) bool {
		return mv.Rating.Average >= 7.5
	})
	for _, mv := range mvs {
		fmt.Println(mv.Title)
	}
}
