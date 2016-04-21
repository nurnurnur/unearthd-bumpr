Unearthd Bumpr
====================================

![](http://puu.sh/8DNw6.jpg)

Overview
--------

This is a go application that builds a playlist of song id's from the JJJ Unearthed website and
"plays" them.

Source Requirements
------------

* GoLang > 1.2.1
* [The Internet!](https://s-media-cache-ak0.pinimg.com/736x/03/5b/48/035b486b37463ddd99945c891eb7f439.jpg)

Pre-compiled Requirements
------------

* [The Internet!](https://s-media-cache-ak0.pinimg.com/736x/03/5b/48/035b486b37463ddd99945c891eb7f439.jpg)

Using Unearthd Bumpr From Source
--------------------------------

* `go run unearthd-bumpr.go` (You'll be asked to enter song id's one by one)
* `go run unearthd-bumpr.go --tracks=123,213,431` (Creates and plays a playlist of Track 123, Track 213 & Track 431)
* `go run unearthd-bumpr.go --file=track_ids.txt` (Text file should be 1 track_id per line)
* `go run unearthd-bumpr.go --playlist=123445` (Plays a JJJ Unearthed playlist)

Using Pre-compiled Packages
---------------------------

Download pre-compiled packages from [DropBox][dropbox]

* `unearthd-bumpr` (You'll be asked to enter song id's one by one)
* `unearthd-bumpr --tracks=123,213,431` (Creates and plays a playlist of Track 123, Track 213 & Track 431)
* `unearthd-bumpr --file=track_ids.txt` (Text file should be 1 track_id per line)
* `unearthd-bumpr --playlist=123445` (Plays a JJJ Unearthed playlist)

Meta
----

* Code: `git clone git://github.com/nurnurnur/unearthd-bumpr.git`

[dropbox]: https://www.dropbox.com/sh/lc2mssvxqcn76t4/AABYBHTldZ6eJPh4ZKP_6AZda
