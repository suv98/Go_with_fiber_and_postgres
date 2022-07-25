package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"gorm.io/gorm"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/subrat/go-fiber-postgres/models"
	"github.com/subrat/go-fiber-postgres/storage"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreatBook(context *fiber.Ctx) error {
	book := Book{}

	err := context.BodyParser(&book)

	if err != nil {
		context.Status(http.StatusUnprocessableEntity).JSON(
			&fiber.Map{"massege": "request failed"})
		return err
	}

	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"massege": "could not create book"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{"massege": "book has been added"})
	return nil
}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(bookModels).Error
	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"massege": "could not get books"})
		return err
	}
	context.Status(http.StatusOK).JSON(&fiber.Map{
		"massege": "Book fetched successfully",
		"data":    bookModels,
	})
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")
	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"massege": "id cannot be empty"})
		return nil
	}

	err := r.DB.Delete(bookModel, id)

	if err.Error != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"massege": "could not delete book"})
		return err.Error
	}
	context.Status(http.StatusOK).JSON(
		&fiber.Map{
			"massege": "book deleted successfully",
		})
	return nil
}

func (r *Repository) GetBookByID(context *fiber.Ctx) error {

	id := context.Params("id")
	bookmodel := &models.Books{}

	if id == "" {
		context.Status(http.StatusInternalServerError).JSON(
			&fiber.Map{"massege": "id cannot be empty"})
		return nil
	}

	fmt.Println("the ID is", id)

	err := r.DB.Where("id=?", id).First(bookmodel).Error

	if err != nil {
		context.Status(http.StatusBadRequest).JSON(
			&fiber.Map{"massege": "could not get the book"})
		return err
	}

	context.Status(http.StatusOK).JSON(
		&fiber.Map{
			"massege": "book id successfully matched",
			"data":    bookmodel,
		})
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreatBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	api.Get("/get_book/:id", r.GetBookByID)
	api.Get("/books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err := storage.NewConnection(config)

	if err != nil {
		log.Fatal("could not load the database")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Repository{
		DB: db,
	}

	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8080")

}
