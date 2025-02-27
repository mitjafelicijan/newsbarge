# Converts RSS feeds into single HTML file

Instead of traditional RSS readers this one exports your items from your feeds
into a single HTML file that you can open in browser and read through your news
there.

Features:

- If RSS feeds item link points to youtube, that youtube video gets embeded
  into the story.
- If feed is podcast it will try to add adui player in the story.
- When you click on a story it will remember that you arelady read it in your
  localStorage in browser.
- You don't need to use a server to serve these HTML files. You can open from
  your filesystem.

## Installation

```console
git clone git@github.com:mitjafelicijan/newsbarge.git
go install .

newsbarge -feed-file=/home/m/.feeds.txt -out-dir=/home/m/Downloads -days-span=14
```

> [!NOTE]
> Check example of `feeds.txt` to see how to add feeds and in what format.

I made alias to my bashrc file for easier use.

```console
alias newsbarge='newsbarge -feed-file=/home/m/.feeds.txt -out-dir=/home/m/Downloads -days-span=14'
```

