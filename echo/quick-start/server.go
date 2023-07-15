package main

import (
	"html/template"
	"io"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

/* Handling Request */
type User struct {
	Name  string `json:"name" xml:"name" form:"name" query:"name"`
	Email string `json:"email" xml:"email" form:"email" query:"email"`
}

func main() {
	/* Hello, World */
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.Logger.Fatal(e.Start(":1323"))

	/* Routing */
	e.POST("/users", saveUser)
	e.GET("/users/:id", getUser) // Path Parameters
	e.PUT("/users/:id", updateUser)
	e.DELETE("/users/:id", deleteUser)

	/* Query parameters */
	e.GET("/show", show)

	/* Form application/x-www-form-urlencoded */
	e.POST("/save", save)

	/* Form multipart/form-data */
	e.POST("/saveMultipart", saveMultipart)

	/* Handling Request */
	e.POST("/users", func(c echo.Context) error {
		u := new(User)
		if err := c.Bind(u); err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, u)
		// return c.XML(http.StatusCreated, u)
	})

	/* Static Content */
	e.Static("/static", "static")

	/* Templates */
	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}
	e.Renderer = t
	e.GET("/hello", Hello)

	/* Advanced */
	/* Named route "foobar" */
	e.GET("/something", func(c echo.Context) error {
		return c.Render(http.StatusOK, "template.html", map[string]interface{}{
			"name": "Dolly!",
		})
	}).Name = "foobar"

	/* Middleware */
	/* Root Level Middleware */
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	/* Group Level Middleware */
	g := e.Group("/admin")
	g.Use(middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {
		if username == "joe" && password == "secret" {
			return true, nil
		}
		return false, nil
	}))
}

func saveUser(c echo.Context) error {
	return c.String(http.StatusOK, "save")
}

/* Path Parameters */
func getUser(c echo.Context) error {
	id := c.Param("id")
	return c.String(http.StatusOK, id)
}

func updateUser(c echo.Context) error {
	return c.String(http.StatusOK, "save")
}

func deleteUser(c echo.Context) error {
	return c.String(http.StatusOK, "save")
}

/* Query parameters */
func show(c echo.Context) error {
	team := c.QueryParam("team")
	member := c.QueryParam("member")
	return c.String(http.StatusOK, "team:"+team+", member:"+member)
}

/* Form application/x-www-form-urlencoded */
func save(c echo.Context) error {
	name := c.FormValue("name")
	email := c.FormValue("email")
	return c.String(http.StatusOK, "name:"+name+", email:"+email)
}

/* Form multipart/form-data */
func saveMultipart(c echo.Context) error {
	name := c.FormValue("name")
	avatar, err := c.FormFile("avatar")

	if err != nil {
		return err
	}

	/* Source */
	src, err := avatar.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	/* Destination */
	dst, err := os.Create(avatar.Filename)
	if err != nil {
		return err
	}
	defer dst.Close()

	/* Copy */
	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	return c.HTML(http.StatusOK, "<b>Thank you! "+name+"</b>")
}
