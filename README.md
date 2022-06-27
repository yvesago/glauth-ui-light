# glauth-ui-light

## Description

**glauth-ui-light** is a small golang web app to manage users and groups from the db files of [glauth ldap server](https://github.com/glauth/glauth) for small business, self hosted, labs infrastructure  or raspbery servers.

[![Coverage Status](https://coveralls.io/repos/github/yvesago/glauth-ui-light/badge.svg?branch=main)](https://coveralls.io/github/yvesago/glauth-ui-light?branch=main)

Thanks of hot-reload feature on glauth configuration file change, **glauth-ui-light** can edit and update glauth configuration and manage ldap users.

**glauth-ui-light** update only users and groups. 

All lines after the first ``[[users]]`` in  glauth configuration file will be updated. The use of same model structure than glauth V2 allow to keep manual edition of non managed features as ``capabilities, loginShell, sshkeys, otpsecret, yubikey, includegroups, ...``. Only comments on these lines are lost.

Current glauth experimental feature on ``customattributes`` is lost due to the need of a patched ``toml`` library.


## Main aims

- Static binary with small customisations
- Keep abbility to edit glauth configuration file
- Minimal javascript use


## Current features

- Custom front page text and application name
- Add new localisations by adding new yaml file
- Authentication with current glauth users if ``allowreadssha256`` is set
- Admin right is defined for members of group defined in ``gidadmin``
- Users in group ``gidcanchgpass`` are common users and can change their passwords 
- Users in group ``giduseotp`` can define and use an One Time Password application like Google Authenticator, andOTP, ...



## Missing glauth V2 features

Current glauth experimental feature on ``customattributes`` is lost due to the need of a patched toml library.


## Some overkill features for a self hosted infrastructure

- only register bcrypt passwords
- daily log rotates
- i18n support
- responsive UI
- TOTP management
- Bcrypt tokens to bypass TOTP
- delayed after 4 failed login
- rate requests limiter against brute force attempts
- password strength
- CSRF
- STS, CSP
- standalone SSL or via reverse proxy
- high tests coverage
- windows, macos builds (not tested)


## Limits

Allow users to self change their passwords can create bad concurrent updates of glauth configuration file.

## Install binary

**Download** last release from https://github.com/yvesago/glauth-ui-light/releases


**Create config file :**
```
#######################
# glauth-ui-light.conf

# dbfile: glauth conf file 
#    with watchconfig enabled for hot-reload on conf file change
#    glauth-ui-light need write access
dbfile = "samples-conf/glauth-sample-simple.cfg"

# run on a non privileged port
port = "0.0.0.0:8080"
# When a self hosted ssl reverse proxy is used :
#  port = "127.0.0.1:8080"

# Custom first page texts
appname = "glauth-ui-light"
appdesc = "Manage users and groups for glauth ldap server"

[sec]
  # TODO set random secrets for CSRF token
  csrfrandom = "secret1"

[passpolicy]
  min = 2
  max = 24
  allowreadssha256 = true  # to be set to false when all passwords are bcrypt
  entropy = 60             # optional password constraint


[cfgusers]
  start = 5000           # start with this uid number 
  gidadmin = 5501        # members of this group are admins
  gidcanchgpass = 5500   # members of this group can change their password
  giduseotp = 5501       # members of this group use OTP

[cfggroups]
  start = 5500   # start with this gid number 
```


**Start :**
```
$ ./glauth-ui -c glauth-ui-light.conf &

$ firefox http://localhost:8080/

```

## Install Debian package

Download last deb file from https://github.com/yvesago/glauth-ui-light/releases

```
$ sudo dpkg -i glauth-ui-light_1.2.0-0~static0_amd64.deb

$ systemctl status glauth-ui-light
● glauth-ui-light.service - Glauth web
     Loaded: loaded (/lib/systemd/system/glauth-ui-light.service; enabled; vendor preset: enabled)
     Active: active (running) since Tue 2022-02-15 18:14:54 CET; 2h 20min ago
   Main PID: 119451 (glauth-ui-light)
      Tasks: 7 (limit: 4475)
     Memory: 5.6M
     CGroup: /system.slice/glauth-ui-light.service
             └─119451 /usr/bin/glauth-ui-light -c /etc/glauth-ui/glauth-ui.cfg

# custom config
$ vi /etc/glauth-ui/glauth-ui.cfg

# logs
$ tail -f /var/log/glauth-ui/app.20220208

```

## Usage

**Home**
![Home](img/1-home.png)

**Login**
![Login](img/2-login.png)

**Change password**
![Change password](img/3-changepass.png)

**TOTP**
![TOTP](img/3-otp.png)

**Manage users**
![Users page](img/4-userspage.png)

**Edit user**
![Edit user](img/4-usersedit.png)

**Delete group**
![Delete group](img/5-delgroup.png)

**Responsive**

![Responsive](img/6-responsive.png)
![Responsive 2](img/6-responsive2.png)


## Localisation
``cp locales/tr.yml locales/it-IT/it.yml``

Translate strings and add new local to config file

```
...
[locale]
  lang = "it"
  path = "locales/"
  langs = ["en-US","fr-FR","it-IT"]
...

```


## Build binary

```
$ git clone https://github.com/yvesago/glauth-ui-light.git

$ cd glauth-ui-light

$ make

# to build binary AND debian/ubuntu package
$ make deb

```

Tests:
```
# linter
$ golangci-lint run ./...                                                                                    

# tests
$ go test ./...
?   	glauth-ui-light	[no test files]
ok  	glauth-ui-light/config	  0.284s
ok  	glauth-ui-light/handlers  0.246s
ok  	glauth-ui-light/helpers	  18.048s # failed login tests need 18s
ok  	glauth-ui-light/routes	  0.019s


# test coverage
$  go test -coverprofile=coverage.out ./... 

$ go tool cover -func=coverage.out
...
glauth-ui-light/routes/routes.go:79:	initServer		85.4%
glauth-ui-light/routes/routes.go:182:	SetRoutes		93.9%
glauth-ui-light/routes/routes.go:230:	contains		100.0%
glauth-ui-light/routes/routes.go:239:	setCacheHeaders		100.0%
glauth-ui-light/routes/routes.go:258:	Auth			100.0%
total:					(statements)		95.1%

# html browser output
$ go tool cover -html=coverage.out
```

## Deploy

```
$ scp -pr locales admin@server:/home/app/glauth-ui-light

$ scp build/linux/glauth-ui  admin@server:/home/app/glauth-ui-light
```

## Build debian/ubuntu package

```
$ apt install build-essential quilt

$ git clone https://github.com/yvesago/glauth-ui-light.git
$ cd glauth-ui-light

$ make deb

# view content
$ dpkg-deb -c ../glauth-ui-light_1.2.0-0~static0_amd64.deb

# install
$ sudo dpkg -i ../glauth-ui-light_1.2.0-0~static0_amd64.deb
```

### Code structure:
```
main.go
|-config
    |-glauth-config-v2.go   // glauth users goups models
    |-webconfig.go
    |-user.go     // user methods   
|-helpers
    |-cookie.go   // cookies for session
    |-db.go       // read write data
    |-18n.go      // i18n
    |-sessions.go // manage session
|-handler
    |-global.go   // global var, render
    |-login.go
    |-users.go
    |-userProfile.go
    |-groups.go
|- routes
    |-routes.go   // load template, i18n. Set routes, auth middleware
    |-web         // due to "embed" files 
       |-assets
          |-css
             |-bootstrap.css
             |-....css
          |-js
             |-....js
          |-templates
             |-global
                |-header.tmpl
                |-footer.tmpl
             |-home
                |-index.tmpl
                |-login.tmpl
             |-user
                |-list.tmpl
                |-add.tmpl
                |-edit.tmpl
                |-profile.tmpl
             |-group
                |-list.tmpl
                |-add.tmpl
                |-edit.tmpl
|-locales
    |-en-US
    |-fr-FR
|-samples-conf
    |-webconfig.cfg
    |-glauthconfig.cfg

```


## References
 
https://github.com/demo-apps/go-gin-app

https://etienner.github.io/connexion-deconnexion-avec-gin/

https://www.alexedwards.net/blog/form-validation-and-processing

https://vincent.bernat.ch/en/blog/2019-pragmatic-debian-packaging

https://github.com/wagslane/go-password-validator


## Changelog

Next v.4.x:
  * Fix issue #4 with on password entropy. Thx to KaptinLin
  * Improve UI. Thx to KaptinLin
  * Add german translation. Thx to publicdesert
  * More user fields Unix and Phone Numbers. Thx to loomanss (TODO: config options and tests)

v1.4.2:
  * Add spanish translations. Thx to Iago Sardiña.

v1.4.1:
  * Add optional password strength constraint


v1.4.0:
  * Add app passwords (tokens) to bypass ldap OTP (bcrypt only)
  * fix denied changes on Lock


v1.2.0:
  * Add OTP management
  * tweak UI


v1.0.1:
  * fix keep unchanged old sha256 password
  * fix UI mistakes


v1.0.0:
  * initial release



## Licence

MIT License

Copyright (c) 2022 Yves Agostini

<yves+github@yvesago.net>


