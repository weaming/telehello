package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/weaming/telehello/core"
	"github.com/weaming/telehello/extension"
)

func GetHotMovieText(city string, score float64) ([]string, error) {
	hotMovies, err := extension.GetMovieInTheaters(city)
	if err != nil {
		return nil, err
	}
	mvs := hotMovies.FilterSubject(func(i int, mv extension.MovieSubject) bool {
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
		if admin, ok := core.ChatsMap[core.AdminKey]; ok {
			txt, err := GetHotMovieText("深圳", score)
			if !core.NotifiedErr(err, admin.ID) {
				core.NotifyText(strings.Join(txt, "\n\n"), admin.ID)
			}
			timer := time.NewTimer(time.Minute * delta)
			<-timer.C
		}
	}
}
