package main

import (
	"encoding/json"
	"io/ioutil"
)

const (
	APIBase            = "https://api.douban.com"
	APIMovieSearch     = "/v2/movie/search?q="
	APIMovieInfo       = "/v2/movie/subject/"
	APIMovieInTheaters = "/v2/movie/in_theaters?city="
)

var http_client_douban = NewHTTPClient(30)

type MovieSubject struct {
	Rating struct {
		Max     int     `json:"max"`
		Average float64 `json:"average"`
		Stars   string  `json:"stars"`
		Min     int     `json:"min"`
	} `json:"rating"`
	Genres []string `json:"genres"`
	Title  string   `json:"title"`
	Casts  []struct {
		Alt     string `json:"alt"`
		Avatars struct {
			Small  string `json:"small"`
			Large  string `json:"large"`
			Medium string `json:"medium"`
		} `json:"avatars"`
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"casts"`
	CollectCount  int    `json:"collect_count"`
	OriginalTitle string `json:"original_title"`
	Subtype       string `json:"subtype"`
	Directors     []struct {
		Alt     string `json:"alt"`
		Avatars struct {
			Small  string `json:"small"`
			Large  string `json:"large"`
			Medium string `json:"medium"`
		} `json:"avatars"`
		Name string `json:"name"`
		ID   string `json:"id"`
	} `json:"directors"`
	Year   string `json:"year"`
	Images struct {
		Small  string `json:"small"`
		Large  string `json:"large"`
		Medium string `json:"medium"`
	} `json:"images"`
	Alt string `json:"alt"`
	ID  string `json:"id"`
}

type MovieList struct {
	Count    int            `json:"count"`
	Start    int            `json:"start"`
	Total    int            `json:"total"`
	Subjects []MovieSubject `json:"subjects"`
	Title    string         `json:"title"`
}

func httpGetBase(url string) ([]byte, error) {
	resp, err := http_client_douban.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(resp.Body)
}

func SearchMovie(name string) (*MovieList, error) {
	url := APIBase + APIMovieSearch + name

	body, err := httpGetBase(url)
	if err != nil {
		return nil, err
	}

	movieList := MovieList{}
	json.Unmarshal(body, &movieList)
	return &movieList, nil
}

func GetMovieInfo(objId string) (*MovieSubject, error) {
	url := APIBase + APIMovieInfo + objId

	body, err := httpGetBase(url)
	if err != nil {
		return nil, err
	}

	movie := MovieSubject{}
	json.Unmarshal(body, &movie)
	return &movie, nil
}

func GetMovieInTheaters(city string) (*MovieList, error) {
	url := APIBase + APIMovieInTheaters + city

	body, err := httpGetBase(url)
	if err != nil {
		return nil, err
	}

	movieList := MovieList{}
	json.Unmarshal(body, &movieList)
	return &movieList, nil
}

func (p *MovieList) FilterSubject(f func(int, MovieSubject) bool) []MovieSubject {
	rv := []MovieSubject{}
	for i, v := range p.Subjects {
		if f(i, v) {
			rv = append(rv, v)
		}
	}
	return rv
}
