package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

const baseURL string = "https://api.tumblr.com/v2/blog/"

var (
	blogid string
	apikey string
)

type blogRes struct {
	Response struct {
		Blog blogInfo `json:"blog"`
	} `json:"response"`
}

type blogInfo struct {
	Name  string `json:"name"`
	Posts int    `json:"posts"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

type postsRes struct {
	Response struct {
		Posts []postInfo `json:"posts"`
	} `json:"response"`
}

type postInfo struct {
	Type       string `json:"type"`
	ID         string `json:"id_string"`
	URL        string `json:"post_url"`
	Date       string `json:"post"`
	Notes      int    `json:"note_count"`
	Slug       string `json:"slug"`
	ReblogRoot string `json:"reblogged_root_name"`
}

type section struct {
	From int
	To   int
}

const delta = 500
const limit = 20

func fetch(url string) ([]byte, error) {
	var res *http.Response
	var err error
	res, err = http.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Error: http status %s", res.Status)
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	return body, err
}

func fetchBlog() (blogInfo, error) {
	url := baseURL + blogid + "/info?api_key=" + apikey
	bytes, err := fetch(url)
	var res blogRes
	if err != nil {
		return res.Response.Blog, err
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return res.Response.Blog, err
	}
	return res.Response.Blog, nil
}

func fetchPosts(offset int) error {
	url := baseURL + blogid + "/posts/" + "?notes_info=true&reblog_info=true&offset=" + strconv.Itoa(offset) + "&api_key=" + apikey
	bytes, err := fetch(url)
	var res postsRes
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return err
	}
	for _, post := range res.Response.Posts {
		if post.ReblogRoot == "" && post.Notes > 0 {
			fmt.Println(offset, post.Slug)
		}
	}
	return nil
}

func fetchAllPosts(total int) error {
	sections := makeSections(total)
	fmt.Println(len(sections))
	fmt.Println(sections)

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(len(sections))
	for _, section := range sections {
		fmt.Println(section)
		go func(start int, end int) {
			defer wg.Done()
			for start < end {
				err := fetchPosts(start)
				if err != nil {
					fmt.Println(start)
				}
				start += limit
			}
		}(section.From, section.To)
	}
	wg.Wait()

	end := time.Now()
	fmt.Println(end.Sub(start))

	return nil
}

func makeSections(total int) []section {
	var sections []section
	for start := 0; start <= total; start += delta {
		end := start + delta - 1
		if end >= total {
			end = total
		}
		sections = append(sections, section{start, end})
	}
	return sections
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func main() {
	blogid = os.Getenv("BLOG_IDENTIFIER")
	apikey = os.Getenv("BLOG_API_KEY")
	if blogid == "" || apikey == "" {
		fmt.Fprintln(os.Stderr, "Missing environment variables BLOG_IDENTIFIER / BLOG_API_KEY")
		os.Exit(1)
	}
	fmt.Println("Fetch tumblr score")
	blog, err := fetchBlog()
	checkErr(err)
	fmt.Println(blog)
	err = fetchAllPosts(blog.Posts)
	checkErr(err)
}
