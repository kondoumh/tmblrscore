package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sort"
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
	Date       string `json:"date"`
	Notes      int    `json:"note_count"`
	Slug       string `json:"slug"`
	ReblogRoot string `json:"reblogged_root_name"`
}

type snapShot struct {
	Date  string     `json:"date"`
	Posts []postInfo `json:"posts"`
}

type section struct {
	From int
	To   int
}

const limit = 20
const numM = 4
const timeLayout = "2006-01-02-15:04"

var allPosts []postInfo

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

func fetchPosts(offset int) ([]postInfo, error) {
	url := baseURL + blogid + "/posts/" + "?notes_info=true&reblog_info=true&offset=" + strconv.Itoa(offset) + "&api_key=" + apikey
	bytes, err := fetch(url)
	var res postsRes
	var posts []postInfo
	if err != nil {
		return posts, err
	}
	err = json.Unmarshal(bytes, &res)
	if err != nil {
		return posts, err
	}
	for _, post := range res.Response.Posts {
		if post.ReblogRoot == "" {
			fmt.Println(offset, post.Slug)
			posts = append(posts, post)
		}
	}
	return posts, nil
}

func fetchAllPosts(total int) {
	sections := makeSections(total)
	fmt.Println(sections)

	start := time.Now()

	var wg sync.WaitGroup
	var mtx = &sync.Mutex{}
	wg.Add(len(sections))
	for _, section := range sections {
		fmt.Println(section)
		go func(start int, end int) {
			defer wg.Done()
			for start < end {
				posts, err := fetchPosts(start)
				if err != nil {
					fmt.Println(start)
				}
				mtx.Lock()
				allPosts = append(allPosts, posts...)
				mtx.Unlock()
				start += limit
			}
		}(section.From, section.To)
	}
	wg.Wait()

	end := time.Now()
	fmt.Println(end.Sub(start))
	fmt.Println(len(allPosts))
}

func makeSections(total int) []section {
	delta := total / numM
	fmt.Printf("Delta : %d\n", delta)
	var sec []section
	for start := 0; start <= total; start += delta {
		end := start + delta - 1
		if end >= total {
			end = total
		}
		sec = append(sec, section{start, end})
	}
	sec[numM-1].To = sec[numM].To
	res := sec[0:numM]
	return res
}

func checkErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}

func writePosts() error {
	sort.Slice(allPosts, func(i, j int) bool { return allPosts[i].Date > allPosts[j].Date })
	var snap snapShot
	jst, _ := time.LoadLocation("Asia/Tokyo")
	snap.Date = time.Now().In(jst).Format(timeLayout)
	snap.Posts = allPosts
	data, _ := json.Marshal(snap)

	file, err := os.Create("posts.json")
	if err != nil {
		return err
	}
	defer file.Close()
	var posts bytes.Buffer
	json.Indent(&posts, data, "", " ")
	file.Write(posts.Bytes())
	return nil
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
	fetchAllPosts(blog.Posts)
	err = writePosts()
	checkErr(err)
}
