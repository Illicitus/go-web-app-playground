package main

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"go_playground/controllers"
	"go_playground/middlewares"
	"go_playground/models"
	"go_playground/utils"
	"net/http"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "user"
	password = "password"
	dbname   = "go_playground"
)

func main() {

	// DB
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname,
	)

	services, err := models.NewServices(psqlInfo)
	utils.Must(err)

	defer func(s *models.Services) {
		if err := services.Close(); err != nil {
			utils.Must(err)
		}
	}(services)

	//utils.Must(services.DestructiveReset())
	utils.Must(services.AutoMigrate())

	// Services and routes
	r := mux.NewRouter()

	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.UserService)
	galleriesC := controllers.NewGalleries(services.GalleryService, services.ImageService, r)

	// Middlewares
	requireUserMw := middlewares.RequireUser{UserService: services.UserService}

	// Routes
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")
	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	r.Handle("/galleries", requireUserMw.ApplyFn(galleriesC.Index)).Methods("GET")
	r.Handle("/galleries/new", requireUserMw.Apply(galleriesC.NewView)).Methods("GET")
	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleriesC.Create)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name("show_gallery")
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleriesC.Edit)).Methods("GET").Name("edit_gallery")
	r.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMw.ApplyFn(galleriesC.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images", requireUserMw.ApplyFn(galleriesC.ImageUpload)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", requireUserMw.ApplyFn(galleriesC.ImageDelete)).Methods("POST")

	imageHandler := http.FileServer(http.Dir("./images"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	fmt.Println("Started on :3000...")

	if err := http.ListenAndServe(":3000", r); err != nil {
		panic(err)
	}
}
