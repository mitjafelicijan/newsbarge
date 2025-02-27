# Converts RSS feeds into HTML

Check example of `feeds.txt` to se ehow to add feeds and in what format.

> [!NOTE]
> This will create HTML with items published in the last week.

```console
git clone git@github.com:mitjafelicijan/newsbarge.git
go install .
```

I put my alias to my bashrc file

```console
alias barge='newsbarge -feed-file=/home/m/.feeds.txt -out-dir=/home/m/Downloads'
```
