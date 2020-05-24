# tmblrscore
![Go](https://github.com/kondoumh/tmblrscore/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/kondoumh/tmblrscore)](https://goreportcard.com/report/github.com/kondoumh/tmblrscore)

Fetch your Tumblr Score(Reblog, like) via Tumblr API.

## Installing

```
go get -u github.com/kondoumh/tmblrscore
```

or 

```
curl -LO https://github.com/kondoumh/tmblrscore/releases/download/<version>/tmblrscore-<platform>-amd64.tar.gz
tar xvf tmblrscore-<platform>-amd64.tar.gz
sudo mv tmblrscore /usr/local/bin
```

## Usage

```
export BLOG_IDENTIFIER=<your.tumblr.blog>
export BLOG_API_KEY=<your.api.key>
tmblrscore
```

JSON file named `posts.json` will be created. 

Content of the JSON is something like

```json
[
 {
  "type": "quote",
  "id_string": "618548293032607744",
  "post_url": "https://reblog.kondoumh.com/post/618548293032607744/",
  "date": "2020-05-19 13:07:00 GMT",
  "note_count": 0,
  "slug": "パパって仮面ライダーなの",
  "reblogged_root_name": ""
 },
 {
  "type": "quote",
  "id_string": "618547857992056832",
  "post_url": "https://reblog.kondoumh.com/post/618547857992056832/",
  "date": "2020-05-19 13:00:05 GMT",
  "note_count": 0,
  "slug": "java-ideの選び方-機能いらんけど使いやすくて安定したのがいい-intellij-idea",
  "reblogged_root_name": ""
 }
]
```