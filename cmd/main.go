package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"slices"
	"time"

	"github.com/labstack/echo/v4"
	// "github.com/labstack/echo/v4/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Defines a "model" that we can use to communicate with the
// frontend or the database
// More on these "tags" like `bson:"_id,omitempty"`: https://go.dev/wiki/Well-known-struct-tags
type BookStore struct {
	MongoID     primitive.ObjectID `bson:"_id,omitempty"`
	ID          string             `bson:"id"      json:"id"`
    BookName    string             `bson:"title"    json:"title"`       
    BookAuthor  string             `bson:"author"  json:"author"`     
    BookEdition string             `bson:"edition" json:"edition"`    
    BookPages   string                `bson:"pages"   json:"pages"`      
    BookYear    string                `bson:"year"    json:"year"`             
}

// Wraps the "Template" struct to associate a necessary method
// to determine the rendering procedure
type Template struct {
	tmpl *template.Template
}

// Preload the available templates for the view folder.
// This builds a local "database" of all available "blocks"
// to render upon request, i.e., replace the respective
// variable or expression.
// For more on templating, visit https://jinja.palletsprojects.com/en/3.0.x/templates/
// to get to know more about templating
// You can also read Golang's documentation on their templating
// https://pkg.go.dev/text/template
func loadTemplates() *Template {
	return &Template{
		tmpl: template.Must(template.ParseGlob("views/*.html")),
	}
}

// Method definition of the required "Render" to be passed for the Rendering
// engine.
// Contraire to method declaration, such syntax defines methods for a given
// struct. "Interfaces" and "structs" can have methods associated with it.
// The difference lies that interfaces declare methods whether struct only
// implement them, i.e., only define them. Such differentiation is important
// for a compiler to ensure types provide implementations of such methods.
func (t *Template) Render(w io.Writer, name string, data interface{}, ctx echo.Context) error {
	return t.tmpl.ExecuteTemplate(w, name, data)
}

// Here we make sure the connection to the database is correct and initial
// configurations exists. Otherwise, we create the proper database and collection
// we will store the data.
// To ensure correct management of the collection, we create a return a
// reference to the collection to always be used. Make sure if you create other
// files, that you pass the proper value to ensure communication with the
// database
// More on what bson means: https://www.mongodb.com/docs/drivers/go/current/fundamentals/bson/
func prepareDatabase(client *mongo.Client, dbName string, collecName string) (*mongo.Collection, error) {
	db := client.Database(dbName)

	names, err := db.ListCollectionNames(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, err
	}
	if !slices.Contains(names, collecName) {
		cmd := bson.D{{"create", collecName}}
		var result bson.M
		if err = db.RunCommand(context.TODO(), cmd).Decode(&result); err != nil {
			log.Fatal(err)
			return nil, err
		}
	}

	coll := db.Collection(collecName)
	return coll, nil
}

// Here we prepare some fictional data and we insert it into the database
// the first time we connect to it. Otherwise, we check if it already exists.
func prepareData(client *mongo.Client, coll *mongo.Collection) {
	startData := []BookStore{
		{
			ID:          "example1",
			BookName:    "The Vortex",
			BookAuthor:  "José Eustasio Rivera",
			BookEdition: "958-30-0804-4",
			BookPages:   "292",
			BookYear:    "1924",
		},
		{
			ID:          "example2",
			BookName:    "Frankenstein",
			BookAuthor:  "Mary Shelley",
			BookEdition: "978-3-649-64609-9",
			BookPages:   "280",
			BookYear:    "1818",
		},
		{
			ID:          "example3",
			BookName:    "The Black Cat",
			BookAuthor:  "Edgar Allan Poe",
			BookEdition: "978-3-99168-238-7",
			BookPages:   "280",
			BookYear:    "1843",
		},
	}

	// This syntax helps us iterate over arrays. It behaves similar to Python
	// However, range always returns a tuple: (idx, elem). You can ignore the idx
	// by using _.
	// In the topic of function returns: sadly, there is no standard on return types from function. Most functions
	// return a tuple with (res, err), but this is not granted. Some functions
	// might return a ret value that includes res and the err, others might have
	// an out parameter.
	for _, book := range startData {
		cursor, err := coll.Find(context.TODO(), book)
		var results []BookStore
		if err = cursor.All(context.TODO(), &results); err != nil {
			panic(err)
		}
		if len(results) > 1 {
			log.Fatal("more records were found")
		} else if len(results) == 0 {
			result, err := coll.InsertOne(context.TODO(), book)
			if err != nil {
				panic(err)
			} else {
				fmt.Printf("%+v\n", result)
			}

		} else {
			for _, res := range results {
				cursor.Decode(&res)
				fmt.Printf("%+v\n", res)
			}
		}
	}
}

// Generic method to perform "SELECT * FROM BOOKS" (if this was SQL, which
// it is not :D ), and then we convert it into an array of map. In Golang, you
// define a map by writing map[<key type>]<value type>{<key>:<value>}.
// interface{} is a special type in Golang, basically a wildcard...
func findAllBooks(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"id":          res.ID,
			"title":    res.BookName,
			"author":  res.BookAuthor,
			"edition": res.BookEdition,
			"pages":   res.BookPages,
		})
	}

	return ret
}

// API Search
func findAllBooksApi(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	var ret []map[string]interface{}
	for _, res := range results {
		ret = append(ret, map[string]interface{}{
			"id":      res.ID,
			"title":    res.BookName,
			"author":  res.BookAuthor,
			"pages":   res.BookPages,
			"edition": res.BookEdition,
			"year": res.BookYear,
		})
	}

	return ret
}

func findAllAuthors(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	authorsM := make(map[string]int)
	var ret []map[string]interface{}

	for _, res := range results {
        authorsM[res.BookAuthor]++
    }

	for author, count := range authorsM {
		var id string
		for _, res := range results {
			if res.BookAuthor == author {
                id = res.MongoID.Hex()
                break
            }
		}

		ret = append(ret, map[string]interface{}{
			"id":          id,
			"author":  author,
			"amountbooks":   count,
		})
		
	}

	return ret
}



func findAllYears(coll *mongo.Collection) []map[string]interface{} {
	cursor, err := coll.Find(context.TODO(), bson.D{{}})
	var results []BookStore
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	yearsM := make(map[string]bool)
	var ret []map[string]interface{}

	for _, res := range results {
		if _, exists := yearsM[res.BookYear]; !exists {
			yearsM[res.BookYear] = true

			ret = append(ret, map[string]interface{}{
				"id":        res.MongoID.Hex(),
				"year":  res.BookYear,
			})
		}
	}

	return ret
}




func main() {
	// fmt.Println("Station 0")

	// Connect to the database. Such defer keywords are used once the local
	// context returns; for this case, the local context is the main function
	// By user defer function, we make sure we don't leave connections
	// dangling despite the program crashing. Isn't this nice? :D
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// fmt.Println("Station 1")

	// TODO: make sure to pass the proper username, password, and port
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))

	// fmt.Println("Station 2")

	// This is another way to specify the call of a function. You can define inline
	// functions (or anonymous functions, similar to the behavior in Python)
	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	// You can use such name for the database and collection, or come up with
	// one by yourself!
	coll, err := prepareDatabase(client, "exercise-1", "information")

	prepareData(client, coll)

	// Here we prepare the server
	e := echo.New()

	// Define our custom renderer
	e.Renderer = loadTemplates()

	// Log the requests. Please have a look at echo's documentation on more
	// middleware
	// e.Use(middleware.Logger())

	e.Static("/css", "css")

	// Endpoint definition. Here, we divided into two groups: top-level routes
	// starting with /, which usually serve webpages. For our RESTful endpoints,
	// we prefix the route with /api to indicate more information or resources
	// are available under such route.
	e.GET("/", func(c echo.Context) error {
		return c.Render(200, "index", nil)
	})

	e.GET("/books", func(c echo.Context) error {
		books := findAllBooks(coll)
		return c.Render(200, "books", books)
	})

	e.GET("/authors", func(c echo.Context) error {
		authors := findAllAuthors(coll)
		return c.Render(200, "authors", authors)
	})

	e.GET("/years", func(c echo.Context) error {
		years := findAllYears(coll)
		return c.Render(200, "years", years)
	})

	e.GET("/search", func(c echo.Context) error {
		return c.Render(200, "search-bar", nil)
	})

	e.GET("/create", func(c echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	// You will have to expand on the allowed methods for the path
	// `/api/route`, following the common standard.
	// A very good documentation is found here:
	// https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Methods
	// It specifies the expected returned codes for each type of request
	// method.
	e.GET("/api/books", func(c echo.Context) error {
		books := findAllBooksApi(coll)
		return c.JSON(http.StatusOK, books)
	})

	// own POST, Update, Delete Methods -> Malaka lets go

	// Some ideas for Post: https://echo.labstack.com/docs/request and https://developer.mozilla.org/en-US/docs/Web/HTTP/Reference/Methods/POST 
	e.POST("/api/books", func(ctx echo.Context) error {
		bookstore := new(BookStore)

		// https://echo.labstack.com/docs/binding 
		if err := ctx.Bind(bookstore); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// debug
		fmt.Printf("\nPOST: Empfangene Daten - ", bookstore)

		if bookstore.ID == "" {                       
			bookstore.ID = primitive.NewObjectID().Hex()
		}

		

		// check for duplicated
		// bson info: https://pkg.go.dev/go.mongodb.org/mongo-driver/bson 
		filterDup := bson.M{
			"id":      bookstore.ID,
			"title": bookstore.BookName,
			"author": bookstore.BookAuthor,
			"pages": bookstore.BookPages,
			"edition": bookstore.BookEdition,
			"year": bookstore.BookYear,
		}

		// Find duplic. mongo https://www.mongodb.com/docs/manual/reference/method/db.collection.countDocuments/#:~:text=count()%20%2C%20db.-,collection.,documents%20in%20a%20sharded%20cluster.
		n, err := coll.CountDocuments(context.TODO(), filterDup)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if n > 0 {
			return echo.NewHTTPError(http.StatusConflict, "duplicate")
		}

		_, err = coll.InsertOne(context.TODO(), bookstore)

		if err != nil {
        	return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
    	}
		fmt.Printf("\nPOST: DONE ")
		return ctx.JSON(http.StatusCreated, bookstore)
	})


	// PUT
	e.PUT("/api/books/:id", func(ctx echo.Context) error {
		id := ctx.Param("id")

		bookstore := new(BookStore)
		if err := ctx.Bind(bookstore); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		// debug
		fmt.Printf("\nPUT: Empfangene Daten - ", bookstore)

		// Update: https://joshua-etim.medium.com/how-i-update-documents-in-mongodb-with-golang-94485dbe54f7 
		// $set : https://www.mongodb.com/docs/manual/reference/operator/update/set/
		updater := bson.M{"$set": bson.M{}}
		if bookstore.BookName != "" { updater["$set"].(bson.M)["title"] = bookstore.BookName}
		if bookstore.BookAuthor != "" { updater["$set"].(bson.M)["author"] = bookstore.BookAuthor}
		if bookstore.BookEdition != "" { updater["$set"].(bson.M)["edition"] = bookstore.BookEdition}
		if bookstore.BookPages != "" { updater["$set"].(bson.M)["pages"] = bookstore.BookPages}
		if bookstore.BookYear != "" { updater["$set"].(bson.M)["year"] = bookstore.BookYear}

		updateResult, err := coll.UpdateOne(context.TODO(), bson.M{"id": id}, updater)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if updateResult.MatchedCount == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "Not found id")
		}

		fmt.Printf("\nPUT: DONE ")
		return ctx.NoContent(http.StatusOK)
	})

	// DELETE
	e.DELETE("/api/books/:id", func(ctx echo.Context) error {
		id := ctx.Param("id")
		delete_res, _ := coll.DeleteOne(context.TODO(), bson.M{"id": id})

		// debug
		fmt.Printf("\nDELETE: Empfangene Daten - ", id)

		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
		}
		if delete_res.DeletedCount == 0 {
			return echo.NewHTTPError(http.StatusNotFound, "Not found id")
		}

		fmt.Printf("\nDELETE: DONE ")
		return ctx.NoContent(http.StatusOK)
	})


	// We start the server and bind it to port 3030. For future references, this
	// is the application's port and not the external one. For this first exercise,
	// they could be the same if you use a Cloud Provider. If you use ngrok or similar,
	// they might differ.
	// In the submission website for this exercise, you will have to provide the internet-reachable
	// endpoint: http://<host>:<external-port>
	e.Logger.Fatal(e.Start(":3030"))
}
