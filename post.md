# Refactoring Go's Web Server Hello World

The first thing a web developer does when they try out a new language is to make a Hello World web application. The simple example in Go is pretty straightforward, but it can be hard to grow to suit the needs of a larger web application. In this post we'll take the canonical hello world go web app example and refactor it twice into a solution that's much easier to work with in the long run.

# Part 1: Hello World

[Go's net/http docs](http://godoc.org/net/http) include an example for running your first web application:

```go
http.Handle("/foo", fooHandler)

http.HandleFunc("/bar", func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
})

log.Fatal(http.ListenAndServe(":8080", nil))
```

I would call this first way of creating a web application the "functional" way or the "package level" way, because we use package level functions to access a hidden global http server and a hidden global logger instance and we're just using an inline function for our handler.

Let's rewrite this example to make something runnable as a `main` package file that responds to `/foo` and logs `request to foo` using go's log package:

```go
package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/foo", func(w http.ResponseWriter, r *http.Request) {
		log.Println("request to foo")
	})

	http.ListenAndServe(":8080", nil)
}
```

For the purposes of this post, we are going to care about two things that our application does:

1. It handles `/foo`
2. It logs using go's standard `log` package to log `request to foo`

These super simple requirements encapsulate two the two techniques that we're going to be focusing on while we refactor:

1. Mapping paths to handlers (a.k.a. routing)
2. Accessing a shared context (our logger)

So let's talk about growth. As we add more routes and handler functions, our `main` will get really really long. Not only that, it will contain *all the functionality* of our application that we don't extract to a library. So it's going to get big! Second, this http server can only exist compiled into a binary that doesn't contain any other http servers with conflicting routes. Since it uses the global `http` package handler functions, it will share the routing namespace with any other http servers. Likewise, our logger will collide with anyone else using the global logger (this is mainly an issue for settings on the logger, like tagging).

# Part 2: Better Routing with our own Globals

Let's address the problems with the previous solution. First, we will extract the logger into our own package level global variable so we can control creating the instance. This way it will be global to our package, but not conflict with other users of the package level `log`. Second, we'll make our own `http.Server` instance onto which we'll map paths to handler functions. Again, this will remove any conflicts with other package level servers. Let's check it out:

```go
package main

import (
	"log"
	"net/http"
	"os"
)

var (
	logger *log.Logger
)

func main() {
	logger = log.New(os.Stdout, "web ", log.LstdFlags)

	server := &http.Server{
		Addr:    ":8080",
		Handler: routes(),
	}

	server.ListenAndServe()
}

func routes() *http.ServeMux {
	r := http.NewServeMux()

	r.HandleFunc("/foo", foo)

	return r
}

func foo(w http.ResponseWriter, r *http.Request) {
	logger.Println("request to foo")
}
```

OK, so we have a `logger` `var` that we can set in `main` and we can access throughout our package. Now we can tag the logger with `web` and also turn on timestamps. That's a nice improvement!

Next up we have two new functions: `routes` and `foo`. `routes` will give us an `http.ServeMux` that is in charge of mapping paths to functions. So now we have one single place where we handle routing, and it doesn't have any of our implementation. `foo` is an `http.HandlerFunc` compliant function, so it can just focus on doing what it's supposed to do. And it's nice and readable as its own separate function, instead of being inlined.

That leaves us with our new `main`. Here, we initialize our global logger, and we also define our http server. We can set it's port here and then point the `Handler` to the mux we create with `routes`. Then we can call `ListenAndServe()`. All in all, we've split our previous solution into much more distinct components, we have:

1. `main` which actually runs the server
2. `routes` which defines which paths map to which functions
3. `foo` which does our actual web server functionality, in this case, just some logging

Let's talk about growth. When we add a new route, we add one line in our `routes`, and we have to create a new handling function. Additionally, we can even have those handling functions come from other packages. So we can chop up our app into sub-applications based on task. Much nicer!

We've also left the global context for `log` and `http`. That means our app won't get screwed up by any other libraries that try to attach anything to `http` or `log`. We also have better control over our logger and server. Great!

But, we can do even better. The problems we still have here is that this application is still very much a `main` style binary, not a package or library. That makes it hard to share and hard to test. It would also be nice to not rely on global variables at all (our `logger`) and encapsulate them within the web server itself.

# Part 3: Object Oriented to the Rescue

In this refactor, we're going to make our application server a real object so that it can encapsulate its dependencies (like `logger`) and also play nicer with other packages and even be easier to test. Let's check it out:

```go
package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	http.ListenAndServe(":8080", New())
}

func New() http.Handler {
	mux := http.NewServeMux()
	log := log.New(os.Stdout, "web ", log.LstdFlags)
	app := &app{mux, log}

	mux.HandleFunc("/foo", app.foo)

	return app
}

type app struct {
	mux *http.ServeMux
	log *log.Logger
}

func (a *app) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *app) foo(w http.ResponseWriter, r *http.Request) {
	a.log.Println("request to foo")
}
```

Starting with `main` you can see that all it does is start a web server on a port, and it uses `New()` to get a `http.Handler`. If we wanted to, `New()` could easily be a function inside a package, like `myapp.New()`, and we could move all the significant code outside of the `main` package.

`New()` is a function that returns an `http.Handler`. By sticking to this interface and not some custom `Application` interface, we will create a much better Go citizen. It will be understood that the object you get back from our library is going to work well with anything that can work with `http.Handler`s. That means we can wrap them with middleware easily and also test them easily, or even embed them in another web application. Also, it means we don't have to export our `Application` type (since the polite thing to do is export any types returned by exported functions). This keeps our API leaner and clearer.

So, `New()` creates our mux like before, but it also creates our log. Note that we could easily add options to `New()` to change how the logging works, or even take th `log` as an option for doing dependency injection.

We then go on to create an `app` wrapping the mux and log. Next, we do our routing by mapping `/foo` to `app.foo`. The cool think about this is that our handler will now run on `app`, giving us access to our entire web application context.

Let's look at `app` the struct. Notice that this is an unexported struct. By having `New()` return `http.Handler` we can use completely unexported objects to build our app, which lowers our package footprint. `app` simply embeds our mux and logger.

Next up, we have `app`'s `ServeHTTP`. This is the only method we have to implement to satisfy the `http.Handler` interface. And all we have to do is delegate it right to our mux. What we're saying here is that we want our mux to respond to all web requests and then use our routing definitions to then call the function that handles the route.

Finally, we have our humble `foo` which is almost unchanged. The only difference is that it is called on `app` and it logs via the `app`'s logger.

So, let's talk about growth! First of all, we can very easily move our app to it's own package. That's great because we can isolate its dependencies and completely protect its unexported methods. When we grow our app, we don't have to change our `main` at all, we simply add more handlers to the mux in `New()` and map them to functions.

Additionally, as our dependencies grow, we have an obvious place to put them: the `app` struct. We could add a third party api client object, a metrics reporter, a database, and whatever else! It's also clear that these should be initialized in New, and any configuration they need become parameters for `New()`. Most of the time, this becomes an environment variable that `main` reads and sends right to `New()`. It's a great pattern.

Then, when we go to test our app, we can bypass `New()` and construct an `app` struct directly (since tests are part of the package being tested) and then we can inject fakes for our app's dependencies and make sure they are used properly. We can also test each handling function individually.

The bottom line here is we've created our web app as a package with a concise and usable API. We've encapsulated and protected our dependencies and removed them from the global scope. And last but not least, we've created clear places to add code when more functionality is added. This helps the code stay clean in the future.
