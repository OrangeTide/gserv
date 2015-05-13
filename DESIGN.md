# gserv Design

## Client to Server protocl

* 'login' - send in login authorization
* 'register' - request new account

login
	username
	login hash ((salt+password)+nonce)

register - enters registration process (accepts subset of commands)
	NONE

say - 
	message

tell -
	target
	message
	
emote -
	message (with special escapes/macros)	

inventory - queries inventory

set - sets an option	
	option
	value

## Server to Client protocol

* 'authreq' - require client to send 'login' or 'register' command as the next valid command
* 'welcome' - a welcome message

authreq:
	NONE
	
loginsalt:
	username - username (not canonical)
	password salt - salt from requested username account (!!!)
	nonce
	
	(if account does not exist, invent a fake response)
	
logincomplete - informs the user the login succeeded
	username - the canonical username this time
	
status - update the status message
	message - it's freeform text
