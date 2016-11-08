# Impressive

![](http://i.imgur.com/dXS1iob.png)

A library for translating UCC's Publish2 calendars into iCal files for use in Google Calendar or other calendar apps. This is the base libary for Impressive - a CLI can be found [here](https://github.com/colm2/impressive-cli) and a web app is [here](https://github.com/colm2/impressive-server).

## Using impressive-cli
To create an iCal file for your college timetable:

* Make sure your timetable is set up on [mytimetable.ucc.ie](https://mytimetable.ucc.ie)
* Have Go installed and on your PATH.
* `go install github.com/colm2/impressive-cli`
* `impressive-cli email@umail.ucc.ie password > file.ical`
