# TeleShell 

TeleShell is a utility application aimed at staying in touch with your machine without exposing any of its TCP ports to the outer world. Why do that if you can just talk to local shell on your machine in a Telegram chat?

For the time being this is more of a fun project rather than a reliable and production ready tool. However it does its job (with some caveats - see below).

## How it works

TeleShell is basically a very simple telegram bot. On your command it fires up bash session on the machine where the bot is running, redirects input from Telegram chat to that session and copies output to the same chat. 

When starting shell TeleShell sources a bash script (**$HOME/.teleshell_setup.sh** by default) that tries to configure the environment so that you don't see much garbage in the chat.

It can also execute single command by directly starting a new process **without** launching shell (e.g. some script or program monitoring health of your machine). 

## Why?

1. No need to keep listening TCP ports open.
2. No need for public IP address.
3. A shell in a Telegram chat right in my phone? Cool! :)

## How to try/install
Currently TeleShell does not have any install scripts etc. (and I doubt it ever will). 
Prerequisites:
* go compiler (I'm on Go 1.16.4, earlier versions should be ok)
* unix-like OS (tested on Linux and macOS)
* bash installed in your system
* Telegram bot dedicated to this task (TeleShell will need a bot token - see below on how to obtain it).

To install:

1. Clone the repository: 

**git clone --depth=1 https://github.com/dmfed/teleshell** 

2. Build the binary: 

**cd teleshell/cmd && go build**

Run **./cmd --help** to make sure that the binary has built and can be executed. 

3. Copy **cmd** to where you want it:

**sudo cp cmd /usr/local/bin/teleshell**

3. Edit a setup script suitable for you (might require some experiments).

A very simple setup script **setup.sh** is contained in the repo. It should be OK to try TeleShell with it. Setup stript gets sourced whenever you start a shell session from Telegram chat. It is aimed at setting up the environment so that you don't see much garbage in the output.

Just feed the script to TeleShell with **-onstart** command-line option OR **cp setup.sh ~/.teleshell_setup.sh** (the default location for it).

NOTE1: sourcing setup script is optional. If no setup script was set up TeleShell will still run.

NOTE2: your regular .bashrc, .bash_aliases etc. get sourced as usual (as if you started a terminal locally). 

4. Edit a config file.

Bot token and you Telegram username are required for TeleShell to operate. Telegram username is for authentication purpose. Telegram usernames are unique and the bot will not talk to anyone else but you. Token is required for the bot to talk to Telegram.

Put your Telegram username and bot token into the sample config file **teleshell.conf**. Invoke TeleShell with **-c** command-line option to make TeleShell use your script OR  **cp teleshell.conf ~/.teleshell.conf** (the default location for config).

Alternatively use environment variables:

**export TELESHELL_TOKEN='your bot token'**

**export TELESHELL_USERNAME='your Telegram username'**

If on start TeleShell finds that these variables are set it will disregard configuration file. 

If you have no experience with Telegram bots - see below on how to register a new bot and obtain a token.

## Obtaining bot token

Go to your Telegram app of choice and talk to [**@BotFather**](https://t.me/BotFather) (the official Telegram account handling bots creation). Follow the instructions (issue "/help" command to get initial directions). Default bot settings (privacy mode on, no inline etc.) are OK to start.

## Using the bot

From this point on you're good to go. Just lauch teleshell with **teleshell** if you copied executable binary to **/usr/local/bin** as suggested in Try/Istall section above above, then go to Telegram app and start a chat with your bot. The bot should show available command on start. Issue */help* command any time to see what the bot can do.

## Caveats

1. Since TeleShell just redirects raw input to/from tty you might notice ANSI escape codes appearing in the chat with TeleShell bot. The way to deal with this is to carefully tailor your setup script being sourced when shell starts. Buy default TeleShell sets **TERM=vt220** and issues **stty -echo**. Other settings in the default setup script are also quite obvious. You might want to add more aliases etc. for your particular machine.

2. Avoid running interactive applications. For exaple TeleShell will handle the sudo password promt just fine. However if you launch vim and see garbage in the chat you'll have to guess yourself that you need to ussue ":q" in the chat to quit vim :)

3. More to come...

## Launching TeleShell when system starts

**cmd** directory in the repo has a sample systemd unit-file **teleshell.service**. Edit it to put in your login name (don't mix up with Telegram name - you need your unix username in the unit file) so that the unit it started on behalf of your selected user.

Put **teleshell.service** into **/etc/systemd/system** (or elsewhere where systemd will notice it). 

Then:

**sudo systemctl daemon-reload**
**sudo systemctl reenable teleshell**

## Credits
First prize for this app actually goes to the author of this library: https://github.com/tucnak/telebot

**telebot** is a very complete and stable implementation of Telegram API in Go. 


## That's all folks! 
Thank you for reading to far!

