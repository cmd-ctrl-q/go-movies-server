package main

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/cmd-ctrl-q/go-movies-server/models"
	"github.com/julienschmidt/httprouter"
)

// send back request OK
type jsonResponse struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

func (app *application) getOneMovie(w http.ResponseWriter, r *http.Request) {
	// get movie id
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.logger.Println(errors.New("invalid id parameter"))
		app.errorJSON(w, http.StatusBadRequest, err)
		return
	}

	movie, err := app.models.DB.Get(id)
	if err != nil {
		app.logger.Println("error getting a movie from db")
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, movie, "movie")
	if err != nil {
		app.logger.Println(errors.New("error marshaling data"))
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) getAllMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := app.models.DB.All()
	if err != nil {
		app.logger.Println("error getting movies from db")
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, movies, "movies")
	if err != nil {
		app.logger.Println("error marshalling data")
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) deleteMovie(w http.ResponseWriter, r *http.Request) {
	// get movie id
	params := httprouter.ParamsFromContext(r.Context())
	id, err := strconv.Atoi(params.ByName("id"))
	if err != nil {
		app.logger.Println("error converting string id to int")
		app.errorJSON(w, http.StatusBadRequest, err)
		return
	}

	// delete movie from db
	err = app.models.DB.DeleteMovie(id)
	if err != nil {
		app.logger.Println("error deleting a movie")
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}

	ok := jsonResponse{
		OK: true,
	}

	err = app.writeJSON(w, http.StatusOK, ok, "response")
	if err != nil {
		app.logger.Println("error marshalling json response")
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}
}

type MoviePayload struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Year        string `json:"year"`
	ReleaseDate string `json:"release_date"`
	Runtime     string `json:"runtime"`
	Rating      string `json:"rating"`
	MPAARating  string `json:"mpaa_rating"`
}

func (app *application) editMovie(w http.ResponseWriter, r *http.Request) {
	var payload MoviePayload

	// read json in request
	err := json.NewDecoder(r.Body).Decode(&payload)
	if err != nil {
		app.logger.Println("error decoding movie: %w", err)
		app.errorJSON(w, http.StatusBadRequest, err)
		return
	}

	var movie models.Movie

	// movie already exists in db
	if payload.ID != "0" {
		id, err := strconv.Atoi(payload.ID)
		if err != nil {
			app.logger.Println("error converting payload.ID to int")
			app.errorJSON(w, http.StatusBadRequest, err)
			return
		}
		m, err := app.models.DB.Get(id)
		if err != nil {
			app.logger.Println("error getting movie from db")
			app.errorJSON(w, http.StatusInternalServerError, err)
		}
		movie = *m
		movie.UpdatedAt = time.Now()
	}

	movie.ID, err = strconv.Atoi(payload.ID)
	if err != nil {
		app.logger.Println("error converting movie.ID to int")
		app.errorJSON(w, http.StatusBadRequest, err)
		return
	}
	movie.Title = payload.Title
	movie.Description = payload.Description
	movie.ReleaseDate, err = time.Parse("2006-01-02", payload.ReleaseDate)
	if payload.ReleaseDate != "" && err != nil {
		app.logger.Println("error parsing movie.ReleaseDate to date")
		app.errorJSON(w, http.StatusBadRequest, err)
		return
	}
	movie.Year = movie.ReleaseDate.Year()
	movie.Runtime, err = strconv.Atoi(payload.Runtime)
	if err != nil {
		app.logger.Println("error converting movie.Runtime to int")
		app.errorJSON(w, http.StatusBadRequest, err)
		return
	}
	movie.Rating, err = strconv.Atoi(payload.Rating)
	if err != nil {
		app.logger.Println("error converting movie.Rating to int")
		app.errorJSON(w, http.StatusBadRequest, err)
		return
	}
	movie.MPAARating = payload.MPAARating
	movie.CreatedAt = time.Now()
	movie.UpdatedAt = time.Now()

	if movie.Poster == "" {
		movie = getPoster(movie)
	}

	// check if movie should be inserted or updated into db
	if movie.ID == 0 {
		// store in db
		err = app.models.DB.InsertMovie(movie)
		if err != nil {
			app.logger.Println("error inserting movie to database")
			app.errorJSON(w, http.StatusInternalServerError, err)
			return
		}
	} else {
		err = app.models.DB.UpdateMovie(movie)
		if err != nil {
			app.logger.Println("error updating movie in database")
			app.errorJSON(w, http.StatusInternalServerError, err)
			return
		}
	}

	ok := jsonResponse{
		OK:      true,
		Message: "Movie edited successfully",
	}

	err = app.writeJSON(w, http.StatusOK, ok, "response")
	if err != nil {
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) getAllGenres(w http.ResponseWriter, r *http.Request) {

	genres, err := app.models.DB.GenresAll()
	if err != nil {
		app.logger.Println(err)
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, genres, "genres")
	if err != nil {
		app.logger.Println(err)
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}
}

func (app *application) getAllMoviesByGenre(w http.ResponseWriter, r *http.Request) {
	// get genre
	params := httprouter.ParamsFromContext(r.Context())
	genreID, err := strconv.Atoi(params.ByName("genre_id"))
	if err != nil {
		app.logger.Println("invalid id parameter")
		app.errorJSON(w, http.StatusBadRequest, err)
		return
	}

	movies, err := app.models.DB.All(genreID)
	if err != nil {
		app.logger.Println("error getting movies from db")
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, movies, "movies")
	if err != nil {
		app.logger.Println("error marshalling data")
		app.errorJSON(w, http.StatusInternalServerError, err)
		return
	}
}

func getPoster(movie models.Movie) models.Movie {
	type TheMovieDB struct {
		Page    int `json:"page"`
		Results []struct {
			Adult            bool    `json:"adult"`
			BackdropPath     string  `json:"backdrop_path"`
			GenreIds         []int   `json:"genre_ids"`
			ID               int     `json:"id"`
			OriginalLanguage string  `json:"original_language"`
			OriginalTitle    string  `json:"original_title"`
			Overview         string  `json:"overview"`
			Popularity       float64 `json:"popularity"`
			PosterPath       string  `json:"poster_path"`
			ReleaseDate      string  `json:"release_date"`
			Title            string  `json:"title"`
			Video            bool    `json:"video"`
			VoteAverage      float64 `json:"vote_average"`
			VoteCount        int     `json:"vote_count"`
		} `json:"results"`
		TotalPages   int `json:"total_pages"`
		TotalResults int `json:"total_results"`
	}

	client := &http.Client{}
	// add api key
	key := os.Getenv("THEMOVIEDB_API_KEY")
	if key == "" {
		log.Fatal("THEMOVIEDB_API_KEY env variable is empty")
	}
	theUrl := "https://api.themoviedb.org/3/search/movie?api_key="
	log.Println(theUrl + key + "&query=" + url.QueryEscape(movie.Title))

	// make request
	req, err := http.NewRequest("GET", theUrl+key+"&query="+url.QueryEscape(movie.Title), nil)
	if err != nil {
		log.Println(err)
		return movie
	}

	// add header
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return movie
	}
	defer resp.Body.Close()

	// get bytes from body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return movie
	}

	var responseObject TheMovieDB

	// unmarshal json
	json.Unmarshal(bodyBytes, &responseObject)

	if len(responseObject.Results) > 0 {
		movie.Poster = responseObject.Results[0].PosterPath
	}

	return movie
}
