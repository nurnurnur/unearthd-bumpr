Unearthd Bumpr
====================================

![](http://puu.sh/8DNw6.jpg)

Overview
--------

This is a go application that builds a playlist of song id's from the JJJ Unearthed website and
"plays" them.

CAVEAT: It only works with tracks uploaded before roughly March 2014

Requirements
------------

* GoLang > 1.2.1
* The Internet

Using Unearthd Bumpr From Source
--------------------------------

* `go run unearthd-bumpr` (You'll be asked to enter song id's one by one)
* `go run unearthd-bumpr 123 213 431` (Creates and plays a playlist of Track 123, Track 213 & Track 431)
* `go run unearthd-bumpr < track_ids.txt` (Text file should be 1 track_id per line)

Using Pre-compiled Packages
---------------------------

Download pre-compiled packages from [DropBox][dropbox]

* `unearthd-bumpr` (You'll be asked to enter song id's one by one)
* `unearthd-bumpr 123 213 431` (Creates and plays a playlist of Track 123, Track 213 & Track 431)
* `unearthd-bumpr < track_ids.txt` (Text file should be 1 track_id per line)

Meta
----

* Code: `git clone git://github.com/nurnurnur/unearthd-bumpr.git`

[dropbox]: https://www.dropbox.com/sh/lc2mssvxqcn76t4/AABYBHTldZ6eJPh4ZKP_6AZda