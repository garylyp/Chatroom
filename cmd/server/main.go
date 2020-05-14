package main

import (
	"io"
	"log"
	"net/http"

	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Types used in this package
type (
	user struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
)

// Global variables
var (
	users = map[int]*user{}
	seq   = 1
	names = [...]string{"Alexis", "Becca", "Cindy"}
)

//----------
// Handlers
//----------

func createUser(c echo.Context) error {
	u := &user{
		ID:   seq,
		Name: names[seq%3],
	}
	if err := c.Bind(u); err != nil {
		return err
	}
	// Potential Race Condition
	users[u.ID] = u
	seq++

	return c.JSON(http.StatusCreated, u)
}

func getUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	return c.JSON(http.StatusOK, users[id])
}

func updateUser(c echo.Context) error {
	u := new(user)
	// Will be able to set customized name later
	if err := c.Bind(u); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	users[id].Name = u.Name
	return c.JSON(http.StatusOK, users[id])
}

func deleteUser(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	delete(users, id)
	return c.NoContent(http.StatusNoContent)
}

func main() {
	// Hello world, the web server
	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		io.WriteString(w, "Hello, world!\n")
	}

	http.HandleFunc("/hello", helloHandler)
	log.Println("Listing for requests at http://localhost:8000/hello")
	// log.Fatal(http.ListenAndServe(":8000", nil))

	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.POST("/users", createUser)
	e.GET("/users/:id", getUser)
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))

}
