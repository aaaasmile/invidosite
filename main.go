package main

import (
	"flag"
	"fmt"
	"invido-site/src/conf"
	"invido-site/src/idl"
	"invido-site/src/watch"
	"log"
	"os"
)

// Example edit a post:
//
// go run .\main.go  -editpost -date "2023-01-04"
//
// Example new post:
//
//	go run .\main.go  -newpost "Quo Vadis" -date "2023-01-04" -watch
//
// # Example edit page
//
// go run .\main.go  -editpage -name "autore"
// Example new page:
//
// go run .\main.go  -newpage "statistiche" -date "2025-01-18" -watch
// Example rebuild all (use it when templates are changed)
// go run .\main.go  -rebuildall
// Example build only changed posts
// go run .\main.go  -buildposts
// Example build only pages (all pages)
// go run .\main.go  -buildpages
// Build the main index page
// go run .\main.go  -buildmain
// Scan and update post info in db
// go run .\main.go  -scancontent
// after edit a new post, prepare everything for the upload
// go run .\main.go -forsync
// buildonly one page
// go run .\main.go  -buildonepage -name "briscola"
func main() {
	var ver = flag.Bool("ver", false, "Prints the current version")
	var configfile = flag.String("config", "config.toml", "Configuration file path")
	var watchdir = flag.Bool("watch", false, "Watch the mdhtml file and generate the html")
	var newpost = flag.String("newpost", "", "title of the new post")
	var date = flag.String("date", "", "Date of the post, e.g. 2025-09-30")
	var editpost = flag.Bool("editpost", false, "edit post at date")
	var editpage = flag.Bool("editpage", false, "edit page at name")
	var newpage = flag.String("newpage", "", "name of the new page")
	var name = flag.String("name", "", "name of the page")
	var rebuildall = flag.Bool("rebuildall", false, "force to create all htmls (links main, post and pages)")
	var scancontent = flag.Bool("scancontent", false, "fill the db table with the missed source content")
	var buildposts = flag.Bool("buildposts", false, "create posts (only changed)")
	var buildpages = flag.Bool("buildpages", false, "create pages (all)")
	var buildonepage = flag.Bool("buildonepage", false, "build one page")
	var buildmain = flag.Bool("buildmain", false, "create main index.html")
	var buildfeed = flag.Bool("buildfeed", false, "create feed.xml")
	var buildtags = flag.Bool("buildtags", false, "create tags html")
	var force = flag.Bool("force", false, "force flag")
	var debug = flag.Bool("debug", false, "debug flag")
	var all4sync = flag.Bool("all4sync", false, "flag to prepare all stuff for the sync")
	flag.Parse()

	if *ver {
		fmt.Printf("%s, version: %s", idl.Appname, idl.Buildnr)
		os.Exit(0)
	}
	if _, err := conf.ReadConfig(*configfile, `../../cert`); err != nil {
		log.Fatal("ERROR: ", err)
	}
	if *scancontent {
		if err := watch.ScanContent(*force, *debug); err != nil {
			log.Fatal("ERROR: ", err)
		}
		return
	} else if *buildmain {
		if err := watch.BuildMain(); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *buildpages {
		if err := watch.BuildPages(*force); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *buildonepage {
		if err := watch.BuildOnePage(*name); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *buildposts {
		if err := watch.BuildPosts(); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *rebuildall {
		if err := watch.RebuildAll(); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *editpost {
		if err := watch.EditPost(*date); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *newpost != "" {
		if err := watch.NewPost(*newpost, *date, *watchdir); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *newpage != "" {
		if err := watch.NewPage(*newpage, *date, *watchdir); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *editpage {
		if err := watch.EditPage(*name); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *buildfeed {
		if err := watch.BuildFeed(); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *buildtags {
		if err := watch.BuildTags(); err != nil {
			log.Fatal("ERROR: ", err)
		}
	} else if *all4sync {
		if err := watch.PrepareForRsync(*debug); err != nil {
			log.Fatal("ERROR: ", err)
		}
	}
	log.Println("That' all folks!")
}
