package main

import (
	"fmt"
	"strings"
	"time"
)

func GetHotMovieText(city string, score float64) ([]string, error) {
	hotMovies, err := GetMovieInTheaters(city)
	if err != nil {
		return nil, err
	}
	mvs := hotMovies.FilterSubject(func(i int, mv MovieSubject) bool {
		return mv.Rating.Average >= score
	})

	var texts []string
	for i, mv := range mvs {
		texts = append(texts,
			fmt.Sprintf("%d 《%v》 [%v分]:\n%v\n%v", i+1,
				mv.Title, mv.Rating.Average, strings.Join(mv.Genres, ", "), mv.Alt))
	}
	return texts, nil
}

func ScanDoubanMovie(score float64, delta time.Duration) {
	for {
		txt, err := GetHotMovieText("深圳", score)
		if !NotifyErr(err) {
			NotifyText(strings.Join(txt, "\n\n"))
		}
		timer := time.NewTimer(time.Minute * delta)
		<-timer.C
	}
}
