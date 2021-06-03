#!/usr/bin/bash
# This is teleshell's onstart script. 
# Its purpose is to prepare shell session to talk to you via Telegram chat: 
# disable terminal echo, disable colors if possible (since color ASCII escapes 
# display as garbage in Telegram. If unsure - just use the defaults here.
#  
# Please copy this script to $HOME/.teleshell_setup.sh so it gets sources on
# every start of teleshel. 
# You can also run teleshell -onstart <yourscript> to source script from 
# different location.
#
# Note that teleshell issues 'stty -echo' command immediately after starting
# shell. If you'd like to have echo on add 'stty echo' below.
export PS1="[teleshell]:\u@\h:\w > "
unset PAGER 
alias systemctl='systemctl --no-pager'
unalias ls
unalias grep

