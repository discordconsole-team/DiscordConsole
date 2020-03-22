/*
DiscordConsole is a software aiming to give you full control over accounts, bots and webhooks!
Copyright (C) 2020 Mnpn

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/
// TRANSLATORS:
// - English, jD91mZM2 & Mnpn
// - Swedish, Mnpn
// - Spanish, ArceCreeper
package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/jD91mZM2/stdutil"
)

var errLangCorrupt = errors.New("corrupt language file")
var lang map[string]string

// TL stands for TransLate kek
func tl(name string) string {
	str, ok := lang[name]
	if ok {
		return str
	}

	return name
}

func loadLangAuto(langfile string) {
	fmt.Println("Loading language...")
	switch langfile {
	case "en":
		loadLangDefault()
	case "sv":
		loadLangString(langSv)
	case "es":
		loadLangString(langEs)
	default:
		reader, err := os.Open(langfile)
		if err != nil {
			stdutil.PrintErr("Could not read language file", err)
			return
		}
		defer reader.Close()

		err = loadLang(reader)
		if err != nil {
			stdutil.PrintErr("Could not load language file", err)
			loadLangDefault()
		}
	}
}
func loadLang(reader io.Reader) error {
	lang = make(map[string]string)
	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {
		text := scanner.Text()
		if text == "" {
			continue
		}
		parts := strings.SplitN(text, "=", 2)
		if len(parts) != 2 {
			return errLangCorrupt
		}
		key := parts[0]
		val := parts[1]

		if strings.HasSuffix(key, ".dev") && devVersion {
			key = key[:len(key)-len(".dev")]
		}

		lang[key] = val
	}

	return scanner.Err()
}
func loadLangString(lang string) error {
	return loadLang(strings.NewReader(lang))
}
func loadLangDefault() {
	loadLangString(langEn)
}

// Here is just some long data.
// This comment is a separator, btw.

// English by jD91mZM2 & Mnpn
var langEn = `
update.checking=Checking for updates...
update.error=Error checking for updates
update.available=Update available! Version
update.available.dev=Latest stable release:
update.download=Download from
update.none=No updates found.

loading.bookmarks=Reading bookmarks...

failed.auth=Couldn't authenticate
failed.avatar=Couldn't set avatar
failed.ban.create=Could not ban user
failed.ban.delete=Could not unban user
failed.ban.list=Could not list bans
failed.base64=Couldn't convert to Base64
failed.block=Couldn't block user
failed.category.create=Could not create category
failed.channel.create=Could not create channel
failed.channel.delete=Could not delete channel
failed.channel=Could not query channel
failed.exec=Could not execute
failed.file.delete=Could not delete file
failed.file.load=Could not load file
failed.file.open=Couldn't open file
failed.file.read=Could not read file
failed.file.save=Could not save file
failed.file.write=Could not write file
failed.fixpath=Could not 'fix' filepath
failed.friend.add=Couldn't add friend
failed.friend.list=Couldn't get friends
failed.friend.remove=Couldn't remove friend
failed.generic=Failed
failed.guild.create=Could not create guild
failed.guild.delete=Could not delete guild
failed.guild.edit=Could not edit guild
failed.guild=Could not query guild
failed.invite.accept=Could not accept invite
failed.invite.create=Invite could not be created
failed.invite=Could not query invite
failed.json=Could not parse json
failed.kick=Could not kick user
failed.leave=Could not leave
failed.lua.event=Recovered from LUA error
failed.lua.run=Could not run lua
failed.members=Could not list members
failed.move=Could not move the user
failed.msg.delete=Couldn't delete message
failed.msg.edit=Couldn't edit message
failed.msg.query=Could not get message
failed.msg.send=Could not send message
failed.nick=Could not set nickname
failed.note=Could not set note
failed.nochannel=Server does not have a channel
failed.paste=Failed to paste clipboard:
failed.path.home=Could not determine value of ~
failed.permcalc=Could not open PermCalc (permission calculator)
failed.perms=No permissions to perform this action.
failed.pin=Couldn't pin the message
failed.react.del=Could not remove reaction
failed.react.delall=Could not delete all reactions
failed.react.used=Emoji used already, skipping
failed.react=Could not react to message
failed.reading=Could not read
failed.readline.read=Could not read line
failed.readline.start=Could not start readline library
failed.revoke=Could not revoke
failed.role.change=Could not add/remove role
failed.role.create=Could not create role
failed.role.delete=Could not delete role
failed.role.edit=Could not edit role
failed.roles=Could not get roles
failed.session.start=Could not open session
failed.settings=Could not query user settings
failed.status=Could not set status
failed.timestamp=Couldn't parse timestamp
failed.transfer=Couldn't transfer ownership
failed.typing=Couldn't start typing
failed.unpin=Couldn't unpin the message
failed.user.edit=Couldn't edit user data
failed.user=Couldn't query user
failed.voice.connect=Could not connect to voice channel
failed.voice.disconnect=Could not disconnect
failed.voice.regions=Could not get regions
failed.voice.speak=Could not start speaking
failed.webrequest=Could not make web request

information.aborted=Aborted.
information.category=Category 
information.created.successfully= was created successfully with ID
information.confirmation=Are you really sure you want to do this?
information.deleted.successfully= was deleted successfully.
information.give.ownership=This will give server ownership to
information.guild=Guild 
information.channel=Channel 
information.irreversible=This action is irreversible!
information.length=Some text was cut short. To see the full text, run
information.moved=Moved the user.
information.note=Note set!
information.revoked.successfully=Revoked
information.wait=Wait a second!
information.warning=Warning!

intro.exit=Press Ctrl+D or type 'exit' to exit.
intro.help=Write 'help' for help.

invalid.bookmark=Bookmark doesn't exist
invalid.cache=No cache available!
invalid.channel.voice=No voice channel selected!
invalid.channel=No channel selected!
invalid.command=Unknown command:
invalid.command2=Do 'help' for help.
invalid.dm=This command doesn't work in a DM.
invalid.guild=No guild selected!
invalid.id=You need to select something to get its ID!
invalid.limit.message=Message exceeds character limit
invalid.music.playing=Already playing something
invalid.not.owner=You're not the server owner!
invalid.number=Not a number
invalid.onlyfor.bots=This command only works for bot users.
invalid.onlyfor.users=This only works for users.
invalid.role=No role with that ID
invalid.source.terminal=You must be in terminal to do this.
invalid.status.offline=The offline status exists, but cannot be set through the API
invalid.substitute=Invalid substitute
invalid.unmatched.quote=Unmatched quote in input string
invalid.value=No such value
invalid.webhook.command=Not an allowed webhook command
invalid.webhook=Webhook format invalid. Format: id/token
invalid.yn=Please type either 'y' or 'n'.

login.finish=Logged in with user ID
login.hidden=[Hidden]
login.starting=Authenticating...
login.token.user=User tokens are prefixed with 'user '
login.token.webhook=Webhook tokens are prefixed with 'webhook ', and their URL or id/token
login.token=Please paste your 'token' here.

pointer.private=Private
pointer.unknown=Unknown

status.avatar=Avatar set!
status.cache=Message cached!
status.channel=Selected channel with ID
status.invite.accept=Accepted invite.
status.invite.create=Created invite with code:
status.loading=Loading...
status.msg.create=Created message with ID
status.msg.delall=Deleted # messages
status.name=Name set!
status.status=Status set!

rl.session=Restarting session...
rl.cache.loc=Reloading location cache...
rl.cache.vars=Deleting cache variables...
`

// Swedish by Mnpn
var langSv = `
update.checking=Letar efter uppdateringar...
update.error=Ett fel inträffade under letandet efter nya uppdateringar.
update.available=En uppdatering finns tillgänglig! Version
update.available.dev=Senast stabila version:
update.download=Ladda ner från
update.none=Hittade inga uppdateringar.

loading.bookmarks=Läser bokmärken...

failed.auth=Kunde inte autentisera
failed.avatar=Kunde inte sätta avatar
failed.ban.create=Kunde inte bannlysa användaren
failed.ban.delete=Kunde inte avbannlysa användaren
failed.ban.list=Kunde inte lista bannlysningar
failed.base64=Kunde inte konvertera till Bas64
failed.block=Kunde inte blockera användaren
failed.category.create=Kunde inte skapa kategori
failed.channel.create=Kunde inte skapa kanal
failed.channel.delete=Kunde inte ta bort kanal
failed.channel=Kunde inte fråga efter kanal
failed.exec=Kunde inte köra
failed.file.delete=Kunde inte ta bort fil
failed.file.load=Kunde inte ladda fil
failed.file.open=Kunde inte öppna fil
failed.file.read=Kunde inte läsa fil
failed.file.save=Kunde inte spara fil
failed.file.write=Kunde inte skriva till
failed.fixpath=Kunde inte 'fixa' sökväg
failed.friend.add=Kunde inte acceptera vän
failed.friend.list=Kunde inte få vänner :(
failed.friend.remove=Kunde inte ta bort vän
failed.generic=Misslyckades
failed.guild.create=Kunde inte skapa server
failed.guild.delete=Kunde inte ta bort server
failed.guild.edit=Kunde inte skapa servern
failed.guild=Kunde inte fråga efter server
failed.invite.accept=Kunde inte acceptera inbjudningen
failed.invite.create=Inbjudningen kunde inte skapas
failed.json=Kunde inte tolka JSON
failed.kick=Kunde inte sparka användaren
failed.leave=Kunde inte lämna
failed.lua.event=Återhämtade från LUA-fel
failed.lua.run=Kunde inte köra lua
failed.members=Kunde inte visa medlemmarna
failed.move=Kunde inte flytta användaren
failed.msg.delete=Kunde inte ta bort meddelande
failed.msg.edit=Kunde inte ändra meddelande
failed.msg.query=Kunde inte ta emot meddelande
failed.msg.send=Kunde inte skicka meddelande
failed.nick=Kunde inte sätta smeknamn
failed.note=Kunde inte sätta anteckning
failed.nochannel=Servern har ingen kanal
failed.paste=Kunde inte klista in:
failed.path.home=Det gick inte att bedöma värdet av ~
failed.permcalc=Kunde inte öppna PermCalc
failed.perms=Inga behörigheter för att utföra den här åtgärden.
failed.pin=Kunde inte fästa meddelandet
failed.react.del=Kunde inte ta bort reaktionen
failed.react.delall=Kunde inte ta bort alla reaktioner
failed.react.used=Emoji redan använd, hoppar
failed.react=Kunde inte reagera
failed.reading=Kunde inte läsa
failed.readline.read=Kunde inte läsa linje
failed.readline.start=Kunde inte starta "readline"-biblioteket
failed.revoke=Kunde inte återkalla
failed.role.change=Kunde inte lägga till/ta bort roll
failed.role.create=Kunde inte skapa roll
failed.role.delete=Kunde inte ta bort roll
failed.role.edit=Kunde inte ändra roll
failed.roles=Kunde inte ta emot roller
failed.session.start=Kunde inte öppna sessionen
failed.settings=Kunde inte ta emot användarinställningar
failed.status=Kunde inte sätta status
failed.timestamp=Kunde inte tolka tidsstämplar
failed.transfer=Kunde inte överföra ägarskap
failed.typing=Kunde inte börja skriva
failed.unpin=Kunde inte ta bort det fästa meddelandet
failed.user.edit=Kunde inte ändra användarinformationen
failed.user=Kunde inte förfråga efter användarinformation
failed.voice.connect=Kunde inte ansluta till röstkanal
failed.voice.disconnect=Kunde inte koppla ifrån
failed.voice.regions=Kunde inte läsa regioner
failed.voice.speak=Kunde inte börja prata
failed.webrequest=Det gick inte att göra webbegäran

information.aborted=Avbrutet.
information.category=Kategorin 
information.created.successfully= skapades med ID
information.confirmation=Vill du verkligen göra detta?
information.deleted.successfully= togs bort.
information.give.ownership=Detta kommer att ge server-ägarskap till
information.guild=Servern 
information.channel=Kanalen 
information.irreversible=Denna åtgärd kan inte ångras!
information.length=Viss text var klippt. För att se hela texten, kör
information.moved=Flyttade användaren.
information.note=Anteckning satt!
information.revoked.successfully=Återkallade
information.wait=Vänta en sekund!
information.warning=Varning!

intro.exit=Tryck Ctrl+D eller kör 'exit' för att avsluta.
intro.help=Kör 'help' för hjälp.

invalid.bookmark=Bokmärket finns inte
invalid.cache=Ingen cache tillgänglig!
invalid.channel.voice=Ingen röstkanal vald!
invalid.channel=Ingen kanal vald!
invalid.command=Okänt kommando:
invalid.command2=Kör 'help' för att få hjälp.
invalid.dm=Detta kommandot fungerar inte i DMs.
invalid.guild=Ingen server vald!
invalid.id=Du måste välja något för att kunna få dess ID!
invalid.limit.message=Meddelande överskrider teckenbegränsningen
invalid.music.playing=Spelar redan något
invalid.not.owner=Du är inte server-ägaren!
invalid.number=Inte ett nummer
invalid.onlyfor.bots=Detta kommandot fungerar endast för bot-användare.
invalid.onlyfor.users=Detta kommandot fungerar endast för användare
invalid.role=Ingen roll med det ID:t
invalid.source.terminal=Du måste vara i en terminal för att göra detta.
invalid.status.offline=Offline-statusen finns men kan inte ställas in via API:n
invalid.substitute=Ogiltigt utbyte
invalid.unmatched.quote=Omatchat citattecken i inmatningssträngen
invalid.value=Inget sådant värde
invalid.webhook.command=Inte ett tillåtet Webhook-commando
invalid.webhook=Webhook-formatet är ogiltit. Format: id/token
invalid.yn=Vänligen skriv antigen 'y' eller 'n'.

login.finish=Loggade in med användar-ID:t
login.hidden=[Dold]
login.starting=Autentiserar...
login.token.user=Användar-'tokens' har prefixet 'user '
login.token.webhook=Webhook-'tokens' har prefixet 'webhook ', och deras URL eller id/token
login.token=Vänligen klistra in en 'token' här.

pointer.private=Privat
pointer.unknown=Okänd

status.avatar=Avatar satt!
status.cache=Meddelande cache-at!
status.channel=Valde kanal med ID
status.invite.accept=Accepterade inbjudan.
status.invite.create=Skapade en inbjudan med kod:
status.loading=Laddar...
status.msg.create=Skapade meddelande med ID
status.msg.delall=Tor bort # meddelanden
status.name=Namn satt!
status.status=Status satt!

rl.cache.loc=Laddar om plats-cache...
rl.cache.vars=Tar bort cache-variablar...
rl.session=Startar om session...

console.=konsoll.
`

// Spanish by ArceCreeper
var langEs = `
update.checking=Buscando actualizaciones...
update.error=Error al buscar actualizaciones.
update.available=¡Actualización disponible! Versión
update.available.dev=Última release estable:
update.download=Descarga de
update.none=No se han encontrado actualizaciones.

loading.bookmarks=Leyendo marcadores...

failed.auth=No se ha podido autenticar
failed.avatar=No se ha podido establecer el avatar.
failed.ban.create=No se ha podido banear al usuario
failed.ban.delete=No se ha podido desbanear al usuario
failed.ban.list=No se han podido listar los baneos
failed.base64=No se ha podido convertir a Base64
failed.block=No se ha podido bloquear al usuario
failed.category.create=No se pudo crear la categoría
failed.channel.create=No se ha podido crear el canal
failed.channel.delete=No se pudo borrar el canal
failed.channel=No se ha podido consultar el canal
failed.exec=No se ha podido ejecutar
failed.file.delete=No se ha podido borrar el archivo
failed.file.load=No se ha podido cargar el archivo
failed.file.open=No se ha podido abrir el archivo
failed.file.read=No se ha podido leer el archivo
failed.file.save=No se ha podido guardar el archivo
failed.file.write=No se ha podido escribir el archivo
failed.fixpath=No se ha podido 'arreglar' la ruta del archivo
failed.friend.list=No se han podido conseguir amigos
failed.generic=Fallido
failed.guild.create=No se pudo crear el servidor
failed.guild.delete=No se pudo borrar el servidor
failed.guild.edit=No se ha podido editar el servidor
failed.guild=No se ha podido consultar el servidor
failed.invite.accept=No se ha podido aceptar la invitación.
failed.invite.create=No se ha podido crear la invitación.
failed.invite=No se ha podido consultar la invitación
failed.json=No se ha podido analizar json
failed.kick=No se ha podido kickear al usuario
failed.leave=No se ha podido dejar el servidor
failed.lua.event=Recuperado de error de LUA
failed.lua.run=No se ha podido ejecutar lua
failed.members=No se ha podido listar a los miembros
failed.move=No se pudo mover el usuario
failed.msg.delete=No se ha podido borrar el mensaje
failed.msg.edit=No se ha podido editar el mensaje
failed.msg.query=No se ha podido recibir el mensaje
failed.msg.send=No se ha podido enviar el mensaje
failed.nick=No se ha podido establecer el apodo
failed.note=No se pudo cambiar la nota.
failed.nochannel=El servidor no tiene un canal.
failed.paste=Fallo al copiar el contenido del portapapeles:
failed.path.home=No se ha podido determinar el valor de ~
failed.permcalc=No se ha podido abrir PermCalc (calculador de permisos)
failed.perms=No hay permisos para ejecutar esta acción.
failed.pin=No se pudo fijar el mensaje
failed.react.used=Emoji ya usado, omitiendo
failed.react=No se ha podido reaccionar al mensaje
failed.reading=No se ha podido leer
failed.readline.read=No se ha podido leer la línea
failed.readline.start=No se ha podido iniciar la librería de readline
failed.revoke=No se pudo anular
failed.role.change=No se ha podido añadir/quitar el rol
failed.role.create=No se ha podido crear el rol
failed.role.delete=¡No se ha podido borrar el rol!
failed.role.edit=No se ha podido editar el rol
failed.roles=No se han podido conseguir los roles
failed.session.start=No se ha podido abrir sesión.
failed.settings=No se ha podido consultar las opciones del usuario
failed.status=No se ha podido establecer el estatus.
failed.timestamp=No se ha podido analizar marca de tiempo
failed.transfer=No se ha podido transferir la propiedad.
failed.typing=No se ha podido empezar a escribir
failed.unpin=No se pudo retirar el mensaje
failed.user.edit=No se ha podido editar los datos del usuario
failed.user=No se ha podido consultar el usuario
failed.voice.connect=No se ha podido conectar al canal de voz.
failed.voice.disconnect=No se ha podido desconectar
failed.voice.regions=No se han podido obtener las regiones
failed.voice.speak=No se ha podido empezar a hablar
failed.webrequest=No se ha podido hacer la solicitud de web

information.aborted=Abortado.
information.category=Categoría 
information.created.successfully= se creó correctamente con ID
information.confirmation=¿Estás realmente seguro de que quieres hacer esto?
information.deleted.successfully=" se borró correctamente.
information.give.ownership=Esto le dará la propiedad del servidor a
information.guild=Servidor 
information.channel=Canal 
information.irreversible=¡Esta acción es irrreversible!
information.moved=Usuario movido.
information.note=¡Nota cambiada!
information.revoked.successfully=Anulado.
information.wait=¡Espera un segundo!
information.warning=¡Advertencia!

intro.exit=Pulsa Ctrl+D o escribe 'exit' para salir.
intro.help=Escribe 'help' para obtener ayuda

invalid.bookmark=No existe el marcador
invalid.cache=Caché no disponible!
invalid.channel.voice=¡No se ha seleccionado ningún chat de voz!
invalid.channel=¡Ningún canal seleccionado!
invalid.command2=Escribe 'help' para obtener ayuda.
invalid.command=Comando desconocido:
invalid.guild=¡No se ha seleccionado ningún servidor!
invalid.id=¡Necesitas seleccionar algo para obtener su ID!
invalid.limit.message=El mensaje excede el límite de caracteres.
invalid.music.playing=Ya se está reproduciendo algo.
invalid.not.owner=¡No eres el propietario del servidor!
invalid.number=No es un número.
invalid.onlyfor.bots=Este comando solo funciona con usuarios de bots.
invalid.onlyfor.users=Esto solo funciona para usuarios.
invalid.role=No hay un rol con esa ID
invalid.source.terminal=Debes estar en la terminal para hacer esto.
invalid.status.offline=El estatus desconectado existe, pero no se puede enviar a través de la API
invalid.substitute=Sustituto inválido
invalid.unmatched.quote=Comillas sin cerrar en la cadena de caracteres entrante
invalid.value=No existe esa variable
invalid.webhook.command=No es un comando webhook permitido
invalid.webhook=Formato de webhook inválido. Formato: id/token
invalid.yn=Por favor escriba 'y'(sí) o 'n'(no).

login.finish=Sesión iniciado con ID de usuario
login.hidden=[Oculto]
login.starting=Autenticando...
login.token.user=Los tokens de usuario se prefijan con 'user '
login.token.webhook=Los tokens de webhook se prefijan con 'webhook ', y su URL o id/token
login.token=Por favor pega tu 'token' aquí.

pointer.private=Privado
pointer.unknown=Desconocido

status.avatar=¡Avatar establecido!
status.cache=¡Mensage mandado al caché!
status.channel=Canal seleccionado con ID
status.invite.accept=Invitación aceptada.
status.invite.create=Invitación creada con código:
status.loading=Cargando...
status.msg.create=Creado mensaje con ID
status.msg.delall=# mensajes borrados.
status.name=¡Nombre establecido!
status.status=¡Estatus establecido!

rl.cache.loc=Recargando el caché del servidor actual...
rl.cache.vars=Borrando variables del caché...
rl.session=Reiniciando sesión...
`
