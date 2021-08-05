package main

import (
	"encoding/json"
	"errors"
	"net/http"
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

func (app *application) searchMovies(w http.ResponseWriter, r *http.Request) {

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
