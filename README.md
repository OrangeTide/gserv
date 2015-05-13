# gserv

A MUD server written in Go.

## Features

* Web client

## Goals

* Object-oriented - (duck typing really) Objects don't have an explicit type. If an object has the right properties then it can be used in that situation. 
* Easy online creation - People who are not programmers can build for the MUD.
* Persistance - by default the game is persistant.
* Zone events - objects may be tagged and all objects with that tag can be sent custom events. This could be used to implement a zone reset, weather system, etc.
* Admin logs - admin actions are fully logged and reviewable by other admins.
* Menu & Form system - define menu templates for character creation, builder, admin, etc.
* Permission system - Renegade style ACS to control menus, commands, etc.
