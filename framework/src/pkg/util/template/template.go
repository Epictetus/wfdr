// Allows one to transpose data structures to html content through templates, allowing the designers to work at least somewhat seperately from the programmers.
package template

// TODO: Add caching in production configuration, or in both if you can hook in with inotify to automatically update the cache when a file is updated. Look at os/inoify.

import (
	"http"
	"os"
	"fmt"
	"io"
	"log"
	"time"
	"json"
	"strings"
	// Local packages.
	dlog "util/dlog"
	"util/perms"
	"github.com/crazy2be/browser"
	"util/fnotify"
)

// A simple struct that should be passed to each page when you render it. Templates can rely on all the information here being available on all templates (or, at least, the ones rendered by go). Some of these will be set by the library automatically, but can be overridden in custom code if needed.
type PageInfo struct {
  Title string
  Name string // "photos", "main", "events", etc.
  Request *http.Request // The actual request object
  Perms *perms.Permissions
  ModuleName map[string]bool // Used for highlighting current item on navbar. Not required to be defined if you follow the recommended naming for the "Name" field, this library tries to guess the ModuleName based off of that if it is not defined.
  Mobile bool // Is it a mobile browser?
  Object interface{} // Custom-defined data to pass to the template.
}

var moduleName = "unknown"
var templates = make(map[string]*Template)

// Sets the name of the module that is using this library, used to select the directory to search for templates.
// Pretty much useless with the new jailing API, so don't use it anymore.
func SetModuleName(modName string) {
	moduleName = strings.ToUpper(modName[:1])+modName[1:]
	dlog.Println("Set module name to", modName)
}

// Renders a set of data provided by the module into the format specified by the request. Normally, this means that the data is rendered into HTML (mobile vs. desktop is automatically selected), but we may add support for ?alt=json or such at a later point
func Render(c io.Writer, r *http.Request, title, name string, data interface{}) {
	if WouldUseJson(r) {
		enc := json.NewEncoder(c)
		err := enc.Encode(data)
		if err != nil {
			dlog.Println(err)
			fmt.Fprintln(c, err)
		}
		return
	}
	var p PageInfo
	if title == "" {
		title = strings.ToTitle(moduleName)
	}
// 	if moduleName == "unknown" {
// 		log.Println("Warning: Attempting to render a template without moduleName being set! Call SetModuleName during the initialization of your module in order to correct this (in main()).")
// 	}
	p.Title = title
	// Removed the modulename because it's not needed in the new framework. However, this will break things in the old framework. *sigh*...
	p.Name = /*moduleName + "/" + */name
	p.Request = r
	perms := perms.Get(r)
	p.Perms = perms
	p.Object = data
	
	dlog.Println(p)
	
	// TODO: Remove Execute() function and do this here.
	Execute(c, &p)
}

// Generic function for 404 errors.
func Error404(c http.ResponseWriter, r *http.Request, data interface{}) {
	c.WriteHeader(404)
	Render(c, r, "404 - Not Found", "shared/errors/404", data)
}

func Error500(c http.ResponseWriter, r *http.Request, data interface{}) {
	c.WriteHeader(500)
	Render(c, r, "500 - Internal Error", "shared/errors/500", data)
}

// Would this template use mobile if it were rendered?
// Changes to a different template for mobile use when:
//  (1) There is such a template in existance AND
//  (2) There is a mobile browser detected (duh).
func WouldUseMobile(r *http.Request) bool {
	return browser.IsMobile(r)
}

// Would this templace use json if rendered? Generally specified by prepending ?alt=json to the url.
func WouldUseJson(r *http.Request) bool {
	if r.FormValue("alt") == "json" {
		return true
	}
	return false
}

// Takes an io.Writer and a stuct of PageInfo, rendering the template specified by the PageInfo struct onto the io.Writer. Automatically selects mobile vs desktop templates, and returns an error if anything goes wrong. If you want more control, too bad until someone adds it in.
// NOTE: This function is depricated! Use Render() instead!
func Execute(wr io.Writer, data *PageInfo) (e os.Error) {
	if len(data.Name) < 1 {
		e = os.NewError("PageInfo template name not specified!")
		return
	}
	
	data.ModuleName = make(map[string]bool, 1)
	data.ModuleName[moduleName] = true
	
	dlog.Println("ModuleName:", moduleName, "Map:", data.ModuleName)
	
	prefix := "tmpl/desktop/"
	if WouldUseMobile(data.Request) {
		prefix = "tmpl/mobile/"
	}
	
	// TODO: Should add caching here in production configuration.
	startTime := time.Nanoseconds()
	
	//templates[prefix+data.Name], e = ParseFile(prefix+data.Name)
	template, err := GetTemplate(prefix+data.Name)
	
	endTime := time.Nanoseconds()
	deltaTime := endTime - startTime
	fmt.Println("Took", float32(deltaTime)/(1000.0*1000.0*1000.0), "seconds to parse template", data.Name)
	
	//template := templates[prefix+data.Name]
	
	if err != nil {
		fmt.Println("Error parsing template '", data.Name, "':", err)
		return
	}
	
	startTime = time.Nanoseconds()
	
	template.Render(wr, data)
	
	endTime = time.Nanoseconds()
	deltaTime = endTime - startTime
	fmt.Println("Took", float32(deltaTime)/(1000.0*1000.0*1000.0), "seconds to render template", data.Name)
	return
}

var watcher *fnotify.Watcher

func init() {
	var err os.Error
	watcher, err = fnotify.NewWatcher()
	if err != nil {
		log.Fatal("Error initializing inotify!", err)
	}
	go watcher.Watch()
}

func LoadTemplate(path string) os.Error {
	fmt.Println("Loading template:", path)
	template, err := ParseFile(path)
	if err != nil {
		dlog.Println("Error parsing template", path, err)
		return err
	}
	templates[path] = template
	return nil
}

func TemplateModified(path string) {
	dlog.Println("Reloading template:", path)
	LoadTemplate(path)
}

func GetTemplate(path string) (*Template, os.Error) {
	template, ok := templates[path]
	if ok {
		return template, nil
	}
	err := LoadTemplate(path)
	if err != nil {
		return nil, err
	}
	fmt.Println(path)
	watcher.Modified(path, TemplateModified)
	
	return templates[path], nil
}